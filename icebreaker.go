package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

func init() {
	flag.StringVar(&token, "t", "", "Authentication Token")
	flag.Parse()
}

var token string

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateStatus(0, "!break")
}

func streamSound(session *discordgo.Session, guildID, channelID string) (err error) {

	// false stands for whether the bot is muted
	// true stands for whether the bot is deafened
	voiceChat, err := session.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// "test.dca" is a placeholder for the time being
	targetFile := workingDir + "/" + "test.dca"
	fmt.Println(targetFile)

	dgvoice.PlayAudioFile(voiceChat, targetFile, make(chan bool))
	return

}

func joinVoice(session *discordgo.Session, msg *discordgo.MessageCreate) {

	// find originating channel of message
	channel, err := session.State.Channel(msg.ChannelID)
	if err != nil {
		log.Fatal(err)
		fmt.Println("An error occurred while looking for the channel: ", err)
		return
	}

	// find the corresponding guild of that channel
	guild, err := session.State.Guild(channel.GuildID)
	if err != nil {
		log.Fatal(err)
		fmt.Println("An error occurred while looking for the guild: ", err)
		return
	}

	// look for the message sender in current voice channel
	for _, voiceState := range guild.VoiceStates {
		if voiceState.UserID == msg.Author.ID {
			streamSound(session, guild.ID, voiceState.ChannelID)
			if err != nil {
				log.Fatal(err)
				fmt.Println("An error occurred while playing the sound: ", err)
			}
			return
		}
	}
}

func messageCreate(session *discordgo.Session, msg *discordgo.MessageCreate) {

	// ignores messages of the bot itself
	if msg.Author.ID == session.State.User.ID {
		return
	}

	if msg.Content == "!break" {
		joinVoice(session, msg)
	}
}

func main() {

	dc, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
		fmt.Println("An error occurred while creating a Discord session: ", err)
		return
	}

	// wait for incoming messages
	dc.AddHandler(messageCreate)

	// process incoming messages, guild
	// information as well as voice states
	dc.Identify.Intents = discordgo.MakeIntent(
		discordgo.IntentsGuilds |
			discordgo.IntentsGuildMessages |
			discordgo.IntentsGuildVoiceStates)

	// open discord socket
	err = dc.Open()
	if err != nil {
		log.Fatal(err)
		fmt.Println("An error occurred while opening the WebSocket: ", err)
		return
	}

	// don't allow the execution of main() to continue until
	// there's a signal that kills the process
	fmt.Println("Bot online. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// close down the session cleanly.
	dc.Close()
}
