package atomicgo

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

type MessageStruct struct {
	guildID     string
	guildName   string
	guildData   *discordgo.Guild
	channelID   string
	channelName string
	channelData *discordgo.Channel
	authorID    string
	authorNum   string
	authorName  string
	authorData  *discordgo.User
	text        string
	files       []string
}

//MessageCreate整形
func MessageCreateEdit(discord *discordgo.Session, m *discordgo.MessageCreate) (messageData MessageStruct) {
	var err error
	messageData.guildID = m.GuildID
	messageData.guildData, err = discord.Guild(messageData.guildID)
	if err == nil {
		messageData.guildName = messageData.guildData.Name
	} else {
		messageData.guildName = "DirectMessage"
	}
	messageData.channelID = m.ChannelID
	messageData.channelData, _ = discord.Channel(messageData.channelID)
	messageData.channelName = messageData.channelData.Name
	messageData.authorID = m.Author.ID
	messageData.authorNum = m.Author.Discriminator
	messageData.authorName = m.Author.Username
	messageData.authorData = m.Author
	log.Println(messageData.authorData)
	log.Println(m.Author)
	messageData.text = m.Content
	filesURL := ""
	if len(m.Attachments) > 0 {
		filesURL = "Files: \""
		for _, file := range m.Attachments {
			filesURL = filesURL + file.URL + ","
			messageData.files = append(messageData.files, file.URL)
		}
		filesURL = filesURL + "\"  "
	}
	log.Print("Guild:\"" + messageData.guildName + "\"  Channel:\"" + messageData.channelName + "\"  " + filesURL + "<" + messageData.authorName + "#" + messageData.authorNum + ">: " + messageData.text)
	return
}

//Embed送信
func SendEmbed(discord *discordgo.Session, channelID string, embed *discordgo.MessageEmbed) {
	_, err := discord.ChannelMessageSendEmbed(channelID, embed)
	PrintError("send Embed", err)
	return
}

//リアクション追加用
func AddReaction(discord *discordgo.Session, channelID string, messageID string, reaction string) {
	err := discord.MessageReactionAdd(channelID, messageID, reaction)
	PrintError("Failed reaction add", err)
	return
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
	return
}

//スラッシュコマンド作成
func SlashCommandCreate(discord *discordgo.Session, guildID string, commands []*discordgo.ApplicationCommand) error {
	for _, command := range commands {
		_, err := discord.ApplicationCommandCreate(discord.State.User.ID, guildID, command)
		if err != nil {
			return fmt.Errorf("Cannot create '%v' command: %v", command.Name, err)
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
