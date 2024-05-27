package main

import (
	"fmt"

	"github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

type Room struct {
	Name    string
	History []Message
	Users   []User
}

type User struct {
	Session  ssh.Session
	Terminal *term.Terminal
}

type Message struct {
	From    string
	Message string
}

func (r *Room) Enter(sess ssh.Session, term *term.Terminal) {
	user := User{
		Session:  sess,
		Terminal: term,
	}

	r.Users = append(r.Users, user)
	entryMsg := Message{From: r.Name, Message: "Welcome to my room!"}
	err := send(user, entryMsg)
	if err != nil {
		fmt.Println("Error occured sending message: ", err)
		return
	}
	for _, m := range r.History {
		err = send(user, m)
		if err != nil {
			fmt.Println("Unable to send messages in room: ", err)
			continue
		}
	}
}

func (r *Room) Leave(sess ssh.Session) {
	r.Users = removeByUsername(r.Users, sess.User())
}

func (r *Room) SendMessage(from, msg string) {
	msgObj := Message{From: from, Message: msg}
	r.History = append(r.History, msgObj)

	for _, user := range r.Users {
		if user.Session.User() != from {
			err := send(user, msgObj)
			if err != nil {
				fmt.Println("Error occured sending message: ", msg, "To user: ", user)
				continue
			}

		}
	}
}
