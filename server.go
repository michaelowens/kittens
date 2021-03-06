package main

import (
	irc "github.com/fluffle/goirc/client"
	"github.com/fluffle/goirc/state"
	"strconv"
	"time"
)

type Server struct {
	// A unique ID will be given to each server when a goroutine
	// commences for the first time. This is used to identify
	// POST requests from our webinterface.
	Id uint64

	// Nick is a string that defines the nick of the bot for this
	// specific server.
	Nick string `sql:"size:32"`

	// RealName is a string that defines the real name of the bot
	// for this specific server.
	RealName string `sql:"size:255"`

	// Host is a string that defines the host of the bot for this
	// specific server.
	Host string

	// ServerName is a string that defines the name of the server
	// that the bot is connecting to. (eg. freenode)
	ServerName string

	// Network is a string that defines the physical link that is
	// going to be used to connect to.
	Network string

	// Port is a number that defines the port that the bot uses
	// to connect to.
	Port int

	// SSL is set to true if the bot is connecting via SSL, and
	// set to false if the bot is not connecting via SSL.
	Ssl bool

	// Password is a string that is only used if connecting to
	// the network requires a password.
	Password string

	// Enabled is set to true if the bot is currently enabled,
	// and set to false if it is not enabled.
	Enabled bool

	// UserId is a foreign key that references the user that owns
	// this server.
	UserId uint64

	// CreatedAt is a timestamp of when the specific
	// user was created at.
	CreatedAt time.Time

	// UpdatedAt is a timestamp of when the specific
	// user was last updated at.
	UpdatedAt time.Time

	// Channels is a slice of Channel structs that define what channels
	// the bot connects to/owns.
	Channels []*Channel `sql:"-"`

	// Conn is the connection that each bot is using to connect
	// to the server.
	Conn *irc.Conn `sql:"-"`

	// Timestamp is a unix timestamp which will be set to time.Now
	// when the bot connects to the server.
	Timestamp int64 `sql:"-"`

	// Connected is set to true when the bot connects to the
	// server and set to false when it disconnects.
	Connected bool `sql:"-"`
}

// Create
func (s *Server) Create() {
	verbf("Creating bot from server struct: %s", s)

	// Create connection
	s.Conn = irc.Client(&irc.Config{
		Me: &state.Nick{
			Nick:  s.Nick,
			Ident: s.Host,
			Host:  s.Host,
			Name:  s.RealName,
		},
		Server:      s.Network,
		Pass:        s.Password,
		SSL:         s.Ssl,
		PingFreq:    30 * time.Second,
		NewNick:     func(s string) string { return s + "_" },
		Version:     "Kittens IRC",
		QuitMessage: "bye!",
		SplitLen:    450,
		Recover:     (*irc.Conn).LogPanic,
	})

	// Enable state tracking
	s.Conn.EnableStateTracking()

	// Add connect handler
	s.Conn.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			s.Timestamp = time.Now().Unix()
			s.Connected = true
			infof("Connected to %s", s.Network)
			s.JoinChannels()
		})

	// Add disconnect handler
	s.Conn.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			s.Connected = false
			infof("Disconnected from %s", s.Network)
			infof("Reconnecting to %s", s.Network)
			s.Connect()
		})

	// Listen for messages
	s.Conn.HandleFunc(irc.PRIVMSG,
		func(conn *irc.Conn, line *irc.Line) {
			// Show output of line currently
			s.Logging(line)
		})

	// Listen for JOIN 
	s.Conn.HandleFunc(irc.JOIN,
		func(conn *irc.Conn, line *irc.Line) {
			// Create new irc user if it doesn't exist
			db.Table("irc_users").Where("nickname = ? and host = ?", line.Nick, line.Host).Attrs(IrcUser{
				Nickname: line.Nick,
				Host:     line.Host,
				ServerId: s.Id,
			}).FirstOrCreate(&IrcUser{})

			// Get irc user
			var ircuser IrcUser
			db.Table("irc_users").Where("nickname = ? and host = ?", line.Nick, line.Host).First(&ircuser)

			// Create new channel related to irc user if it doesn't exist
			db.Table("irc_user_channels").Where("channel = ? and irc_user_id = ?", line.Args[0], ircuser.Id).Attrs(IrcUserChannel{
				Channel: line.Args[0],
				IrcUserId: ircuser.Id,
				Modes: "",
				LastJoinedAt: time.Now(),
				LastPartedAt: time.Now(),
			}).FirstOrCreate(&IrcUserChannel{})

			// Get the irc user channel
			var iuc IrcUserChannel
			db.Table("irc_user_channels").Where("channel = ? and irc_user_id = ?", line.Args[0], ircuser.Id).First(&iuc)

			// Set the LastJoinedAt time
			iuc.LastJoinedAt = time.Now()
			db.Save(&iuc)
		})

	verbf("Finished creating bot for server %s", s.ServerName)

	// Connect server if enabled
	if s.Enabled {
		s.Connect()
	}
}

// Connect
func (s *Server) Connect() {
	verbf("Beginning to connect to %s", s.Network)

	// Now we connect
	if s.Enabled {
		if err := s.Conn.ConnectTo(s.Network + ":" + strconv.Itoa(s.Port)); err != nil {
			warnf("Error connecting: %s", err)
			info("Retrying in 30 seconds")
			time.Sleep(30 * time.Second)
			s.Connect()
		}
	} else {
		infof("Not connecting to %s because enabled is false", s.ServerName)
	}
}

// Join Channels is a func that is called when a bot connects
// to a server. The func loops over the channels that are in
// the slice of channels in our Server struct.
func (s *Server) JoinChannels() {
	for i := range s.Channels {
		verbf("Joining channel: %s", s.Channels[i].Name)
		s.Conn.Join(s.Channels[i].Name)
	}
}

// Join New Channel is a func that is called when the bot is
// joining one specific channel for the first time.
func (s *Server) JoinNewChannel(channel string) {
	// Create channel
	ch := Channel{
		Name:     channel,
		ServerId: s.Id,
	}

	// Insert channel into database
	db.Create(&ch)

	// Add channel to struct
	s.Channels = append(s.Channels, &ch)

	// Only join channels if connected
	if s.Connected {
		verbf("Joining channel: %s", channel)
		s.Conn.Join(channel)
	}
}
