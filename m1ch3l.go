package main

type EventType int

// Types of Event
const (
	E_PRIVMSG    = iota // A PRIVMSG sent in private or on a channel
	E_NOTICE            // Same as PRIVMSG but for notices
	E_JOIN              // A user joined (origin is the channel, data is the user)
	E_PART              // A user left a chan (same as E_JOIN)
	E_DISCONNECT        // The bot has been disconnected from a server
	E_QUIT              // A user has quit (origin is the user)
	E_NICK              // A user changed nick (origin is old nick, data is the new nick)
	E_PING              // A ping has been sent by the server
	E_KICK              // A user has been kicked (user is the admin, data is the target)
)

type ActionType int

// Types of Action
const (
	A_SAY        = iota // Say something (requires server, channel, data)
	A_KICK              // Kick someone (requires server, channel, user, optionally data)
	A_JOIN              // Join a channel (requires server, channel)
	A_PART              // Leave a channel (requires server, channel, optionally data)
	A_OP                // Op a user (requires server, channel, user)
	A_RAW               // Send a raw IRC command (requires server), not yet implemented
	A_SENDNOTICE        // Send a notice (requires server), not yet implemented
	A_NEWMODULE         // This is sent by a module after it connects for the first time to the bot. Modules should not care about this type

)

// Events are sent by the server to each module
type Event struct {
	AdminCmd bool      // Is this event admin-issued?
	Server   string    // The server on which the event occured
	Channel  string    // The #channel on which the event occured
	User     string    // Nickname of the user who triggered the event
	Type     EventType // Type of the event
	Data     string    // Additional data (may change depending on the event)
	Raw      string    // Raw command
	CmdId    int       // Id of the command
}

// Priority of the action (meaningful when anti-flood protection is enabled)
const (
	PRIORITY_LOW    = 1
	PRIORITY_MEDIUM = 2
	PRIORITY_HIGH   = 3
)

// Actions are sent by modules to perform an action on the server
type Action struct {
	Server   string     // What server to operate on
	Channel  string     // What channel to operate on
	User     string     // Who is concerned
	Data     string     // Additional data
	Priority int        // Priority of the message
	Type     ActionType // What to do
	Raw      string     // If Type = RAW, send this directly over the server
}

func main() {
}