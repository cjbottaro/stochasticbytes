package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	openai "github.com/sashabaranov/go-openai"
)

func main() {

	discord_token := os.Getenv("DISCORD_TOKEN")
	if discord_token == "" {
		panic("DISCORD_TOKEN is required")
	}

	dg, err := discordgo.New("Bot " + discord_token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	openai_token := os.Getenv("OPENAI_TOKEN")
	if openai_token == "" {
		panic("OPENAI_TOKEN is required")
	}

	c := openai.NewClient(openai_token)

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		handleMessage(c, s, m)
	})

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentMessageContent | discordgo.IntentGuildMembers

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func handleMessage(c *openai.Client, s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !includesUser(m.Mentions, s.State.User) {
		return
	}

	s.ChannelTyping(m.ChannelID)

	fmt.Printf("Incoming message: %s\n", m.Content)

	members, err := s.GuildMembers(m.GuildID, "", 100)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Discord is not cooperating: %v\n", err))
	}

	chatlog, err := s.ChannelMessages(m.ChannelID, 50, "", "", "")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Discord is not cooperating: %v\n", err))
	}

	// Reverse the messages so the oldest is first.
	for i, j := 0, len(chatlog)-1; i < j; i, j = i+1, j-1 {
		chatlog[i], chatlog[j] = chatlog[j], chatlog[i]
	}

	// var chatlogHistory []map[string]string
	// for _, message := range chatlog {
	// 	member := findMemberById(members, message.Author.ID)

	// 	chatlogHistory = append(chatlogHistory, map[string]string{
	// 		"Name":     member.Nick,
	// 		"UserID":   message.Author.ID,
	// 		"Username": message.Author.Username,
	// 		"Message":  message.Content,
	// 	})
	// }
	// chatlogJson, _ := json.Marshal(chatlogHistory)

	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: `You are a Discord bot for a Discord server named Superbolide.
								Your name and username is ChatGPT, and you use the OpenAI API. 
								You should refer to people only by their name when responding.
								You should not use username or userid in responses unless specifically asked to.
								You should not refer to the "conversation history" in your responses.
								Answer as if you are the bot that said the things in the chat logs.`,
		},
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "The server members are described by this JSON: " + membersJson(members),
		},
	}

	messages = append(messages, model3(members, chatlog, m, s)...)

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: m.Content,
		Name:    m.Author.Username,
	})

	resp, err := c.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: messages,
	})
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("My head hurts: %v\n", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, resp.Choices[0].Message.Content)
}

func model1(members []*discordgo.Member, chatlog []*discordgo.Message, m *discordgo.MessageCreate, s *discordgo.Session) []openai.ChatCompletionMessage {
	var output []openai.ChatCompletionMessage

	for _, message := range chatlog {
		if message.Author.ID == s.State.User.ID {
			output = append(output, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Name:    "ChatGPT",
				Content: message.Content,
			})
		} else if includesUser(message.Mentions, s.State.User) {
			output = append(output, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Name:    message.Author.Username,
				Content: message.Content,
			})
		} else {
			member := findMemberById(members, message.Author.ID)
			data := map[string]string{
				"Name":     member.Nick,
				"UserID":   message.Author.ID,
				"Username": message.Author.Username,
				"Message":  message.Content,
			}
			json, _ := json.Marshal(data)
			output = append(output, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Name:    message.Author.Username,
				Content: string(json),
			})
		}
	}

	return output
}

func model2(members []*discordgo.Member, chatlog []*discordgo.Message, m *discordgo.MessageCreate, s *discordgo.Session) []openai.ChatCompletionMessage {
	var output []openai.ChatCompletionMessage
	var data []map[string]string
	var nickname string

	for _, message := range chatlog {
		if message.Author.ID == s.State.User.ID {
			nickname = "ChatGPT"
		} else {
			nickname = findMemberById(members, message.Author.ID).Nick
		}

		data = append(data, map[string]string{
			"Name":      nickname,
			"Username":  message.Author.Username,
			"UserID":    message.Author.ID,
			"Message":   message.Content,
			"Timestamp": message.Timestamp.Format(time.RFC3339),
		})
	}

	json, _ := json.Marshal(data)

	output = append(output, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "The chat logs are described this JSON: " + string(json),
	})

	return output
}

func model3(members []*discordgo.Member, chatlog []*discordgo.Message, m *discordgo.MessageCreate, s *discordgo.Session) []openai.ChatCompletionMessage {
	var output []openai.ChatCompletionMessage

	for _, message := range chatlog {
		if message.Author.ID == s.State.User.ID {
			output = append(output, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Name:    "ChatGPT",
				Content: message.Content,
			})
			fmt.Printf("ChatGPT: %s\n\n", message.Content)
		} else if includesUser(message.Mentions, s.State.User) {
			output = append(output, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Name:    message.Author.Username,
				Content: message.Content,
			})
			fmt.Printf("%s to ChatGPT: %s\n\n", message.Author.Username, message.Content)
		} else {
			member := findMemberById(members, message.Author.ID)
			data := map[string]string{
				"Name":     member.Nick,
				"UserID":   message.Author.ID,
				"Username": message.Author.Username,
				"Message":  message.Content,
			}
			json, _ := json.Marshal(data)
			output = append(output, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Name:    message.Author.Username,
				Content: "A chat log entry was made as JSON: " + string(json),
			})
			fmt.Printf("%s: %s\n\n", message.Author.Username, message.Content)
		}
	}

	return output
}

func debug(s *discordgo.Session, m *discordgo.MessageCreate) {
	// members, err := s.GuildMembers(m.GuildID, "", 100)
	// if err != nil {
	// 	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Discord is not cooperating: %v\n", err))
	// }

	// chatlog, err := s.ChannelMessages(m.ChannelID, 50, "", "", "")
	// if err != nil {
	// 	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Discord is not cooperating: %v\n", err))
	// }

	// // Reverse the messages so the oldest is first.
	// for i, j := 0, len(chatlog)-1; i < j; i, j = i+1, j-1 {
	// 	chatlog[i], chatlog[j] = chatlog[j], chatlog[i]
	// }

	// var chatlogHistory []map[string]string
	// for _, message := range chatlog {
	// 	member := findMemberById(members, message.Author.ID)
	// 	var nickname string
	// 	if message.Author.ID == s.State.User.ID {
	// 		nickname = "ChatGPT"
	// 	} else {
	// 		nickname = member.Nick
	// 	}

	// 	chatlogHistory = append(chatlogHistory, map[string]string{
	// 		"Name":     nickname,
	// 		"UserID":   message.Author.ID,
	// 		"Username": message.Author.Username,
	// 		"Message":  message.Content,
	// 	})
	// }
	// chatlogJson, _ := json.Marshal(chatlogHistory)

	// fmt.Printf("chatlogJson: %v\n", string(chatlogJson))
}

func membersJson(members []*discordgo.Member) string {
	var membersInfo []map[string]string
	for _, member := range members {
		membersInfo = append(membersInfo, map[string]string{
			"Name":     member.Nick,
			"UserID":   member.User.ID,
			"Username": member.User.Username,
		})
	}

	payload, _ := json.Marshal(membersInfo)

	return string(payload)
}

func findMemberById(members []*discordgo.Member, id string) *discordgo.Member {
	for _, member := range members {
		if member.User.ID == id {
			return member
		}
	}
	return nil
}

func includesUser(users []*discordgo.User, u *discordgo.User) bool {
	for _, user := range users {
		if user.ID == u.ID {
			return true
		}
	}
	return false
}
