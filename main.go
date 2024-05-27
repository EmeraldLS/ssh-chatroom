package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

func removeByUsername(users []User, userInSess string) []User {
	var index int
	for i, user := range users {
		if user.Session.User() == userInSess {
			index = i
			break
		}
	}

	return append(users[:index], users[:index+1]...)
}

func send(u User, m Message) error {
	raw := m.From + ">" + m.Message + "\n"
	_, err := u.Terminal.Write([]byte(raw))
	return err
}

var (
	sessions       map[ssh.Session]*Room
	availableRooms []*Room
	enterCmd       = regexp.MustCompile(`^/enter.*`)
	helpCmd        = regexp.MustCompile(`^/help.*`)
	exitCmd        = regexp.MustCompile(`^/exit.*`)
	listCmd        = regexp.MustCompile(`^/list.*`)
	createCmd      = regexp.MustCompile(`^/create.*`)
)

func helpMsg() string {
	return `
	Hello and welcome to the chat server! Please use one of the following commands:
	1. /list: To list available rooms
	2. /enter <room>: To enter a room
	3. /exit: To leave the server
	4. /help: To display this message
	5. /create: To create a room
`
}

func filter[T any](s []T, cond func(t T) bool) []T {
	res := []T{}
	for _, v := range s {
		if cond(v) {
			res = append(res, v)
		}
	}
	return res
}

func listRooms() string {
	var sb strings.Builder

	for _, room := range availableRooms {
		_, _ = sb.WriteString(room.Name + "\n")
	}
	return sb.String()
}

func createRoom(name string) {
	room := &Room{Name: name, History: make([]Message, 0), Users: make([]User, 0)}
	availableRooms = append(availableRooms, room)
}

func chat(s ssh.Session) {
	term := term.NewTerminal(s, fmt.Sprintf("%s > ", s.User()))
	for {
		line, err := term.ReadLine()
		if err != nil {
			break
		}
		if len(line) > 0 {
			if string(line[0]) == "/" {
				switch {

				case exitCmd.MatchString(line):
					return

				case createCmd.MatchString(line):
					roomName := strings.Split(line, " ")[1]
					createRoom(roomName)
					term.Write([]byte("Room created successfully. Use /enter <name> to enter the room"))

				case listCmd.MatchString(line):
					_, err = term.Write([]byte(listRooms()))
					if err != nil {
						fmt.Println("Unable to write to terminal: ", err)
					}

				case enterCmd.MatchString(line):
					roomToEnter := strings.Split(line, " ")[1]
					matching := filter(availableRooms, func(r *Room) bool {
						return roomToEnter == r.Name
					})
					if len(matching) == 0 {
						term.Write([]byte("Invalid Room!\n"))
					} else {
						if sessions[s] != nil {
							sessions[s].Leave(s)
						}
						r := matching[0]
						r.Enter(s, term)
						sessions[s] = r
					}

				case helpCmd.MatchString(string(line)):
					term.Write([]byte(helpMsg()))
				default:
					term.Write([]byte((helpMsg())))

				}
			} else {
				if sessions[s] != nil {
					sessions[s].SendMessage(s.User(), line)
				} else {
					term.Write([]byte((helpMsg())))
				}
			}
		}
	}
}

func main() {
	availableRooms = []*Room{
		{
			Name: "Flirt",
		},
		{
			Name: "Gaming",
		},
		{
			Name: "Memes",
		},
	}

	sessions = make(map[ssh.Session]*Room)

	ssh.Handle(func(s ssh.Session) {
		chat(s)
	})

	log.Println("starting ssh server on port 2222...")
	log.Fatal(ssh.ListenAndServe(":2222", nil))
}
