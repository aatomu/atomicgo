package atomicgo

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

type MessageStruct struct {
	GuildID     string
	GuildName   string
	GuildData   *discordgo.Guild
	ChannelID   string
	ChannelName string
	ChannelData *discordgo.Channel
	UserID      string
	UserNum     string
	UserName    string
	UserData    *discordgo.User
	Text        string
	Files       []string
}

type ReactionStruct struct {
	GuildID     string
	GuildName   string
	GuildData   *discordgo.Guild
	ChannelID   string
	ChannelName string
	ChannelData *discordgo.Channel
	UserID      string
	UserNum     string
	UserName    string
	UserData    *discordgo.User
	MessageData *discordgo.Message
	MessageID   string
	Message     string
	Emoji       string
}

//Botを起動 Port占有させないため
//	defer atomicgo.DiscordBotCleanup(discord)
//が必要
func DiscordBotBoot(botToken string) (discord *discordgo.Session) {
	//bot起動準備
	discord, err := discordgo.New()
	PrintError("Failed get discordStruct", err)

	//token入手
	discord.Token = "Bot " + botToken

	//起動
	err = discord.Open()
	PrintError("Failed Login", err)

	return
}

//Botの使用していたws/wssを削除
func DiscordBotCleanup(discord *discordgo.Session) {
	err := discord.Close()
	PrintError("Failed Leave", err)
}

//MessageCreate整形
func MessageViewAndEdit(discord *discordgo.Session, m *discordgo.MessageCreate) (mData MessageStruct) {
	var err error
	mData.GuildID = m.GuildID
	mData.GuildData, err = discord.Guild(mData.GuildID)
	if err == nil {
		mData.GuildName = mData.GuildData.Name
	} else {
		mData.GuildName = "DirectMessage"
	}
	mData.ChannelID = m.ChannelID
	mData.ChannelData, _ = discord.Channel(mData.ChannelID)
	mData.ChannelName = mData.ChannelData.Name
	mData.UserID = m.Author.ID
	mData.UserNum = m.Author.Discriminator
	mData.UserName = m.Author.Username
	mData.UserData = m.Author
	mData.Text = m.Content
	filesURL := ""
	if len(m.Attachments) > 0 {
		filesURL = "Files: \""
		for _, file := range m.Attachments {
			filesURL = filesURL + file.URL + ","
			mData.Files = append(mData.Files, file.URL)
		}
		filesURL = filesURL + "\"  "
	}
	log.Print("Guild:\"" + mData.GuildName + "\"  Channel:\"" + mData.ChannelName + "\"  " + filesURL + "<" + mData.UserName + "#" + mData.UserNum + ">: " + mData.Text)
	return
}

//MessageCreate整形
func ReactionViewAndEdit(discord *discordgo.Session, r *discordgo.MessageReaction) (rData ReactionStruct) {
	var err error
	rData.GuildID = r.GuildID
	rData.GuildData, err = discord.Guild(rData.GuildID)
	if err == nil {
		rData.GuildName = rData.GuildData.Name
	} else {
		rData.GuildName = "DirectMessage"
	}
	rData.ChannelID = r.ChannelID
	rData.ChannelData, _ = discord.Channel(rData.ChannelID)
	rData.ChannelName = rData.ChannelData.Name
	rData.UserID = r.UserID
	rData.UserData, _ = discord.User(r.UserID)
	rData.UserNum = rData.UserData.Discriminator
	rData.UserName = rData.UserData.Username
	rData.Emoji = r.Emoji.Name
	rData.MessageID = r.MessageID
	rData.MessageData, err = discord.ChannelMessage(rData.ChannelID, r.MessageID)
	if err == nil {
		rData.Message = rData.MessageData.Content
	}

	//改行あとを削除
	if strings.Contains(rData.Message, "\n") {
		replace := regexp.MustCompile(`\n.*`)
		rData.Message = replace.ReplaceAllString(rData.Message, "..")
	}

	//文字数を制限
	nowCount := 0
	logText := ""
	for _, word := range strings.Split(rData.Message, "") {
		if nowCount < 20 {
			logText = logText + word
		}
		if nowCount == 20 {
			logText = logText + ".."
		}
		nowCount++
	}

	//ログを表示
	log.Print("Guild:\"" + rData.GuildName + "\"  Channel:\"" + rData.ChannelData.Name + "\" <" + rData.UserName + "#" + rData.UserNum + "> " + rData.Emoji + " => " + logText)
	return
}

//Embed送信
func SendEmbed(discord *discordgo.Session, channelID string, embed *discordgo.MessageEmbed) {
	_, err := discord.ChannelMessageSendEmbed(channelID, embed)
	PrintError("send Embed", err)
}

//リアクション追加用
func AddReaction(discord *discordgo.Session, channelID string, messageID string, reaction string) {
	err := discord.MessageReactionAdd(channelID, messageID, reaction)
	PrintError("Failed reaction add", err)
}

//ユーザーIDからVCに接続
func JoinUserVCchannel(discord *discordgo.Session, userID string) (*discordgo.VoiceConnection, error) {
	vs := UserVCState(discord, userID)
	return discord.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, true)
}

func UserVCState(discord *discordgo.Session, userid string) *discordgo.VoiceState {
	for _, guild := range discord.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userid {
				return vs
			}
		}
	}
	return nil
}

//音再生
func PlayAudioFile(speed float64, pitch float64, vcsession *discordgo.VoiceConnection, filename string) error {
	if err := vcsession.Speaking(true); err != nil {
		return err
	}
	defer vcsession.Speaking(false)

	opts := dca.StdEncodeOptions
	opts.CompressionLevel = 0
	opts.RawOutput = true
	opts.Bitrate = 120
	opts.AudioFilter = fmt.Sprintf("aresample=24000,asetrate=24000*%f/100,atempo=100/%f*%f", pitch*100, pitch*100, speed)
	encodeSession, err := dca.EncodeFile(filename, opts)
	if err != nil {
		return err
	}

	done := make(chan error)
	stream := dca.NewStream(encodeSession, vcsession, done)
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case err := <-done:
			if err != nil && err != io.EOF {
				return err
			}
			encodeSession.Truncate()
			return nil
		case <-ticker.C:
			playbackPosition := stream.PlaybackPosition()
			log.Println("Sending Now... : Playback:", playbackPosition)
		}
	}
}

//所有ロールの確認
//許可されたロールの所有者か確認
func HaveRole(discord *discordgo.Session, guildID string, userID string, checkRole string) bool {
	//ロールを持っていたら実行
	userRoleList, err := discord.GuildMember(guildID, userID)
	if err != nil {
		PrintError("Failed get UserData on Guild", err)
		return false
	}
	guildRoleList, err := discord.GuildRoles(guildID)
	if err != nil {
		PrintError("Failed get GuildRoles", err)
		return false
	}
	//ロール一覧から検索
	for _, guildRole := range guildRoleList {
		if guildRole.ID == checkRole {
			for _, userRole := range userRoleList.Roles {
				if userRole == guildRole.ID {
					return true
				}
			}
		}
	}
	return false
}

//Botのステータスアップデート
func BotStateUpdate(discord *discordgo.Session, gameName string) {
	state := discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: gameName,
				Type: 0,
			},
		},
		AFK:    false,
		Status: "online",
	}
	discord.UpdateStatusComplex(state)
}

//スラッシュコマンド作成
func SlashCommandCreate(discord *discordgo.Session, guildID string, commands []*discordgo.ApplicationCommand) error {
	for _, command := range commands {
		_, err := discord.ApplicationCommandCreate(discord.State.User.ID, guildID, command)
		if err != nil {
			return fmt.Errorf("cannot create '%s' command: %v", command.Name, err)
		}
	}
	return nil
}

//スラッシュコマンドレスポンス送信
func SlashCommandResponse(discord *discordgo.Session, interaction *discordgo.Interaction, resp *discordgo.InteractionResponse) error {
	err := discord.InteractionRespond(interaction, resp)
	if err != nil {
		return err
	}
	return nil
}
