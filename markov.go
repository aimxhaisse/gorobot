package main

type MarkovHandler struct {
	Config			MarkovConfig
}

type MarkovConfig struct {
	PrivateMessages MarkovConfigForEntry
	Targets map[string]MarkovConfigForEntry
}

type MarkovConfigForEntry struct {
	AlwaysRespondToQuestions bool
	Verbosity		 int
}

func Markov(chac chan Action, chev chan Event, config MarkovConfig) {
	for {
		e := <-chev
		if e.Type == E_PRIVMSG {
		}
	}	
}
