package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkoukk/tiktoken-go"
	openai "github.com/sashabaranov/go-openai"
)

type strategyF func(
	[]*discordgo.Member,
	[]*discordgo.Message,
	*discordgo.MessageCreate,
	*discordgo.Session,
) []openai.ChatCompletionMessage

// type jsonM map[string]any

var (
	_strategy         strategyF
	_gptModel         string
	_tt               *tiktoken.Tiktoken
	_maxHistoryTokens int
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

	// Choose our model.
	_gptModel = os.Getenv("MODEL")
	if _gptModel == "" {
		_gptModel = "gpt-3.5-turbo"
	}
	fmt.Printf("Using model %s\n", _gptModel)

	// Choose our strategy.
	{
		strategy := os.Getenv("STRATEGY")
		if strategy == "" {
			strategy = "3"
		}
		switch strategy {
		case "1":
			_strategy = strategy1
		case "2":
			_strategy = strategy2
		case "3": // Default
			_strategy = strategy3
		}
		fmt.Printf("Using strategy %s\n", strategy)
	}

	// Setup our tokenizer.
	_tt, err = tiktoken.EncodingForModel(_gptModel)
	if err != nil {
		panic(err)
	}

	// Set max history tokens.
	{
		maxHistoryTokens := os.Getenv("MAX_HISTORY_TOKENS")
		if maxHistoryTokens == "" {
			_maxHistoryTokens = 2000
		} else {
			_maxHistoryTokens, err = strconv.Atoi(maxHistoryTokens)
			if err != nil {
				panic(err)
			}
		}
		fmt.Printf("Max history tokens set to %d\n", _maxHistoryTokens)
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

	members, err := s.GuildMembers(m.GuildID, "", 100)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Discord is not cooperating: %v\n", err))
		return
	}

	chatlog, err := getMessageHistory(s, m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Discord is not cooperating: %v\n", err))
		return
	}
	chatlog = chatlog[1:]

	// Reverse the messages so the oldest is first.
	for i, j := 0, len(chatlog)-1; i < j; i, j = i+1, j-1 {
		chatlog[i], chatlog[j] = chatlog[j], chatlog[i]
	}

	// The initial prompt.
	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: `You are a Discord bot for a Discord server named Superbolide.
								Server members can be identified by their name, username, or userid.
								You will be provided with member info JSON.
								Your name and username is ChatGPT.
								You should refer to people by their name.
								You should not use username or userid in responses unless specifically asked to.
								You will be provided with the conversation history via JSON chat logs.
								You should not refer to the "conversation history" in your responses.
								You should not refer to the "chat history" in your responses.
								Be brief with your answers unless otherwise instructed.`,
		},
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "The server members are described by this JSON: " + membersJson(members),
		},
	}

	// Make messages describing the conversation so far.
	messages = append(messages, _strategy(members, chatlog, m, s)...)

	// Append the actual user prompt.
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: m.ContentWithMentionsReplaced(),
		Name:    m.Author.Username,
	})

	fmt.Printf("%s: %s\n\n", m.Author.Username, m.ContentWithMentionsReplaced())

	resp, err := c.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: messages,
		// Functions: []openai.FunctionDefinition{
		// 	{
		// 		Name:        "web_search",
		// 		Description: "Search the web (via Google) for current events or information",
		// 		Parameters: jsonM{
		// 			"type": "object",
		// 			"properties": jsonM{
		// 				"query": jsonM{
		// 					"type":        "string",
		// 					"description": "Google search query string",
		// 				},
		// 			},
		// 			"required": []string{"query"},
		// 		},
		// 	},
		// },
	})
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("My head hurts: %v\n", err))
		return
	}

	fmt.Printf("resp: %+v\n\n", resp)

	if resp.Choices[0].FinishReason == openai.FinishReasonFunctionCall {
		msg := fmt.Sprintf("Hold on, I need to use the internet: %s, %s\n",
			resp.Choices[0].Message.FunctionCall.Name,
			resp.Choices[0].Message.FunctionCall.Arguments,
		)
		s.ChannelMessageSend(m.ChannelID, msg)
		return
	}

	s.ChannelMessageSend(m.ChannelID, resp.Choices[0].Message.Content)
}

func strategy1(members []*discordgo.Member, chatlog []*discordgo.Message, m *discordgo.MessageCreate, s *discordgo.Session) []openai.ChatCompletionMessage {
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

func strategy2(members []*discordgo.Member, chatlog []*discordgo.Message, m *discordgo.MessageCreate, s *discordgo.Session) []openai.ChatCompletionMessage {
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

func strategy3(members []*discordgo.Member, chatlog []*discordgo.Message, m *discordgo.MessageCreate, s *discordgo.Session) []openai.ChatCompletionMessage {
	var output []openai.ChatCompletionMessage
	var content string

	for _, message := range chatlog {
		content = message.ContentWithMentionsReplaced()

		fmt.Printf("%s: %s\n\n", message.Author.Username, content)

		if message.Author.ID == s.State.User.ID {
			output = append(output, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Name:    "ChatGPT",
				Content: content,
			})
		} else if includesUser(message.Mentions, s.State.User) {
			output = append(output, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Name:    message.Author.Username,
				Content: content,
			})
		} else {
			member := findMemberById(members, message.Author.ID)
			data := map[string]string{
				"Name":     member.Nick,
				"UserID":   message.Author.ID,
				"Username": message.Author.Username,
				"Message":  content,
			}
			json, _ := json.Marshal(data)
			output = append(output, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Name:    message.Author.Username,
				Content: "A chat log entry was made as JSON: " + string(json),
			})
		}
	}

	return output
}

func getMessageHistory(s *discordgo.Session, channelID string) ([]*discordgo.Message, error) {
	tokens := 0
	lastMessageID := ""
	d := time.Now()
	var messages []*discordgo.Message

	defer func() {
		fmt.Printf("%d messages, %d tokens, %s\n\n",
			len(messages), tokens, time.Since(d).String(),
		)
	}()

	for {
		batch, err := s.ChannelMessages(channelID, 100, lastMessageID, "", "")
		if err != nil {
			return messages, err
		}

		for _, m := range batch {
			count := len(_tt.Encode(m.ContentWithMentionsReplaced(), nil, nil))
			if count+tokens > _maxHistoryTokens {
				return messages, nil
			}
			messages = append(messages, m)
			tokens = count + tokens
			lastMessageID = m.ID
		}
	}
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
