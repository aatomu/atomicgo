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

type MessageData struct {
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
	Message     string
	MessageID   string
	MessageData *discordgo.Message
	Files       []string
	FormatText  string
}

type VoiceStateData struct {
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
	IsJoin      bool
	FormatText  string
}

type ReactionData struct {
	GuildID      string
	GuildName    string
	GuildData    *discordgo.Guild
	ChannelID    string
	ChannelName  string
	ChannelData  *discordgo.Channel
	UserID       string
	UserNum      string
	UserName     string
	UserData     *discordgo.User
	Message      string
	MessageID    string
	MessageData  *discordgo.Message
	ReactionType string
	Emoji        string
	FormatText   string
}

func DiscordBotSetup(botToken string) (discord *discordgo.Session) {
	//bot起動準備
	discord, err := discordgo.New("Bot " + botToken)
	PrintError("Failed get discordStruct", err)
	return
}

// Botを起動 Port占有させないため
//
//	defer atomicgo.DiscordBotEnd(discord)
//
// が必要
func DiscordBotStart(discord *discordgo.Session) {
	//起動
	err := discord.Open()
	PrintError("Failed Login", err)
}

// Botの使用していたws/wssを削除
func DiscordBotEnd(discord *discordgo.Session) {
	err := discord.Close()
	PrintError("Failed Leave", err)
}

// MessageCreate整形
func MessageParse(discord *discordgo.Session, m *discordgo.MessageCreate) (mData MessageData) {
	var err error
	mData.GuildID = m.GuildID
	mData.GuildData, err = discord.Guild(mData.GuildID)
	if err == nil {
		mData.GuildName = mData.GuildData.Name
	}
	mData.ChannelID = m.ChannelID
	mData.ChannelData, err = discord.Channel(mData.ChannelID)
	if err == nil {
		mData.ChannelName = mData.ChannelData.Name
	}
	mData.UserID = m.Author.ID
	mData.UserNum = m.Author.Discriminator
	mData.UserName = m.Author.Username
	mData.UserData = m.Author
	mData.Message = m.Content
	mData.MessageID = m.ID
	mData.MessageData, _ = discord.ChannelMessage(mData.ChannelID, mData.MessageID)
	filesURL := ""
	if len(m.Attachments) > 0 {
		filesURL = "Files: \""
		for _, file := range m.Attachments {
			filesURL = filesURL + file.URL + ","
			mData.Files = append(mData.Files, file.URL)
		}
		filesURL = filesURL + "\"  "
	}

	mData.FormatText = fmt.Sprintf(`Guild:"%s"  Channel:"%s"  %s<%s#%s>: %s`, mData.GuildName, mData.ChannelName, filesURL, mData.UserName, mData.UserNum, mData.Message)
	return
}

// VCupdate
func VoiceStateParse(discord *discordgo.Session, v *discordgo.VoiceStateUpdate) (vData VoiceStateData) {
	var err error
	vData.GuildID = v.GuildID
	vData.GuildData, err = discord.Guild(vData.GuildID)
	if err == nil {
		vData.GuildName = vData.GuildData.Name
	}
	vData.ChannelID = v.ChannelID
	vData.ChannelData, err = discord.Channel(vData.ChannelID)
	if err == nil {
		vData.ChannelName = vData.ChannelData.Name
	}
	vData.UserID = v.UserID
	vData.UserData, err = discord.User(v.UserID)
	if err == nil {
		vData.UserNum = vData.UserData.Discriminator
		vData.UserName = vData.UserData.Username
	} else {
		vData.UserNum = "    "
		vData.UserName = "????"
	}
	vData.IsJoin = (v.ChannelID != "")

	//ログを表示
	vData.FormatText = fmt.Sprintf(`Guild:"%s"  Channel:"%s"  <%s#%s> IsJoin:"%t"`, vData.GuildName, vData.ChannelName, vData.UserName, vData.UserNum, vData.IsJoin)
	return
}

// ReactionAdd整形
// ReactionType: add remove remove_all
func ReactionParse(discord *discordgo.Session, r *discordgo.MessageReaction, reactionType string) (rData ReactionData) {
	var err error
	rData.GuildID = r.GuildID
	rData.GuildData, err = discord.Guild(rData.GuildID)
	if err == nil {
		rData.GuildName = rData.GuildData.Name
	}
	rData.ChannelID = r.ChannelID
	rData.ChannelData, err = discord.Channel(rData.ChannelID)
	if err == nil {
		rData.ChannelName = rData.ChannelData.Name
	}
	rData.UserID = r.UserID
	rData.UserData, err = discord.User(r.UserID)
	if err == nil {
		rData.UserNum = rData.UserData.Discriminator
		rData.UserName = rData.UserData.Username
	} else {
		rData.UserNum = "    "
		rData.UserName = "????"
	}
	rData.ReactionType = reactionType
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
	rData.FormatText = fmt.Sprintf(`Guild:"%s"  Channel:"%s"  <%s#%s> Type:"%s" Emoji:"%s" => <%s#%s> %s`, rData.GuildName, rData.ChannelName, rData.UserName, rData.UserNum, rData.ReactionType, rData.Emoji, rData.MessageData.Author.Username, rData.MessageData.Author.Discriminator, logText)
	return
}

// Embed送信
func SendEmbed(discord *discordgo.Session, channelID string, embed *discordgo.MessageEmbed) {
	_, err := discord.ChannelMessageSendEmbed(channelID, embed)
	PrintError("Failed Send Embed", err)
}

// リアクション追加用
func AddReaction(discord *discordgo.Session, channelID string, messageID string, reaction string) {
	err := discord.MessageReactionAdd(channelID, messageID, reaction)
	PrintError("Failed Reaction add", err)
}

// ユーザーIDからVCに接続
func JoinUserVCchannel(discord *discordgo.Session, userID string, micMute, speakerMute bool) (vc *discordgo.VoiceConnection, err error) {
	vs := UserVCState(discord, userID)
	if vs == nil {
		return nil, fmt.Errorf("user doesn't join voice chat")
	}
	vc, err = discord.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, micMute, speakerMute)
	return vc, err
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

// 音再生
// end := make(<-chan bool, 1)
func PlayAudioFile(speed float64, pitch float64, vcsession *discordgo.VoiceConnection, filename string, isPlayback bool, end <-chan bool) error {
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
	if !isPlayback {
		ticker.Stop()
	}

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
		case <-end:
			encodeSession.Cleanup()
			_, err := stream.Finished()
			if err != nil {
				PrintError("Failed stop audio", err)
			}
			return nil
		}
	}
}

// 所有ロールの確認
// 許可されたロールの所有者か確認
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

// Botのステータスアップデート
func BotStateUpdate(discord *discordgo.Session, gameName string, Type int) (success bool) {
	state := discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: gameName,
				Type: discordgo.ActivityType(Type),
			},
		},
		AFK:    false,
		Status: "online",
	}
	err := discord.UpdateStatusComplex(state)
	return !PrintError("Failed Update State", err)
}
