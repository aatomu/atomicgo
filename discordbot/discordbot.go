package discordbot

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/atomu21263/atomicgo/utils"
	"github.com/bwmarrin/discordgo"
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
	Status       VoiceStatus
	StatusUpdate VoiceStatus
	FormatText   string
}

type VoiceStatus struct {
	ChannelJoin  bool
	ServerDeaf   bool
	ServerMute   bool
	ClientDeaf   bool
	ClientMute   bool
	ClientGoLive bool
	ClientCam    bool
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

// discordbot init
func Init(token string) (discord *discordgo.Session, err error) {
	return discordgo.New("Bot " + token)
}

// discordbot start
// pls "defer discord.Close()"
func Start(discord *discordgo.Session) error {
	return discord.Open()
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

	// Formatter
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
	vData.Status = VoiceStatus{
		ChannelJoin:  (v.ChannelID != ""),
		ServerDeaf:   v.Deaf,
		ServerMute:   v.Mute,
		ClientDeaf:   v.SelfDeaf,
		ClientMute:   v.SelfMute,
		ClientGoLive: v.SelfStream,
		ClientCam:    v.SelfVideo,
	}
	if v.BeforeUpdate == nil {
		vData.StatusUpdate.ChannelJoin = true
	} else {
		vData.StatusUpdate = VoiceStatus{
			ChannelJoin:  (v.ChannelID != v.BeforeUpdate.ChannelID),
			ServerDeaf:   (v.Deaf != v.BeforeUpdate.Deaf),
			ServerMute:   (v.Mute != v.BeforeUpdate.Mute),
			ClientDeaf:   (v.SelfDeaf != v.BeforeUpdate.SelfDeaf),
			ClientMute:   (v.SelfMute != v.BeforeUpdate.SelfMute),
			ClientGoLive: (v.SelfStream != v.BeforeUpdate.SelfStream),
			ClientCam:    (v.SelfVideo != v.BeforeUpdate.SelfVideo),
		}
	}

	// Formatter
	vData.FormatText = fmt.Sprintf(`Guild:"%s"  Channel:"%s"  <%s#%s>`, vData.GuildName, vData.ChannelName, vData.UserName, vData.UserNum)
	switch {
	case vData.StatusUpdate.ChannelJoin:
		vData.FormatText = fmt.Sprintf("%s ChangeTo ChannelJoin:\"%t\"", vData.FormatText, vData.Status.ChannelJoin)
	case vData.StatusUpdate.ServerDeaf:
		vData.FormatText = fmt.Sprintf("%s ChangeTo ServerDeaf:\"%t\"", vData.FormatText, vData.Status.ServerDeaf)
	case vData.StatusUpdate.ServerMute:
		vData.FormatText = fmt.Sprintf("%s ChangeTo ServerMute:\"%t\"", vData.FormatText, vData.Status.ServerMute)
	case vData.StatusUpdate.ClientDeaf:
		vData.FormatText = fmt.Sprintf("%s ChangeTo ClientDeaf:\"%t\"", vData.FormatText, vData.Status.ClientDeaf)
	case vData.StatusUpdate.ClientMute:
		vData.FormatText = fmt.Sprintf("%s ChangeTo ClientMute:\"%t\"", vData.FormatText, vData.Status.ClientMute)
	case vData.StatusUpdate.ClientGoLive:
		vData.FormatText = fmt.Sprintf("%s ChangeTo ClientGoLive:\"%t\"", vData.FormatText, vData.Status.ClientGoLive)
	case vData.StatusUpdate.ClientCam:
		vData.FormatText = fmt.Sprintf("%s ChangeTo ClientCam:\"%t\"", vData.FormatText, vData.Status.ClientCam)
	}
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

	// Delete After New lines
	if strings.Contains(rData.Message, "\n") {
		replace := regexp.MustCompile(`\n.*`)
		rData.Message = replace.ReplaceAllString(rData.Message, "..")
	}

	logText := utils.StrCut(rData.Message, "..", 20)

	// Formatter
	rData.FormatText = fmt.Sprintf(`Guild:"%s"  Channel:"%s"  <%s#%s> Type:"%s" Emoji:"%s" => <%s#%s> %s`, rData.GuildName, rData.ChannelName, rData.UserName, rData.UserNum, rData.ReactionType, rData.Emoji, rData.UserName, rData.UserNum, logText)
	return
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

	done := make(chan error)
	stream := NewFileEncodeStream(vcsession, filename, EncodeOpts{
		Compression: 1,
		AudioFilter: fmt.Sprintf("aresample=24000,asetrate=24000*%f/100,atempo=100/%f*%f", pitch*100, pitch*100, speed),
	}, done)

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
			stream.Cleanup()
			return nil
		case <-ticker.C:
			log.Printf("Sending Now... : Playback:%.2f(x%.2f)", stream.Status.Time.Seconds(), stream.Status.Speed)
		case <-end:
			stream.Cleanup()
			return nil
		}
	}
}

// 所有ロールの確認
// 許可されたロールの所有者か確認
func HaveRole(discord *discordgo.Session, guildID string, userID string, checkRole string) (bool, error) {
	//ロールを持っていたら実行
	userRoleList, err := discord.GuildMember(guildID, userID)
	if err != nil {
		return false, err
	}
	guildRoleList, err := discord.GuildRoles(guildID)
	if err != nil {
		return false, err
	}
	//ロール一覧から検索
	for _, guildRole := range guildRoleList {
		if guildRole.ID == checkRole {
			for _, userRole := range userRoleList.Roles {
				if userRole == guildRole.ID {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// Botのステータスアップデート
func BotStateUpdate(discord *discordgo.Session, gameName string, Type int) (success bool, err error) {
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
	err = discord.UpdateStatusComplex(state)
	return err == nil, err
}
