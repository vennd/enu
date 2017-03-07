package counterpartyhandlers

import "sync"

var counterparty_BackEndPollRate = 2000 // milliseconds

var counterparty_Mutexes = struct {
	sync.RWMutex
	m map[string]*sync.Mutex
}{m: make(map[string]*sync.Mutex)}
