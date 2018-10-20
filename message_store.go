package messages

// Store is a data structure that holds ArborMessages and allows them
// to be easily looked up by their identifiers. It is safe for
// concurrent use.
type Store struct {
	m        map[string]*Message
	add      chan *Message
	request  chan string
	response chan *Message
}

// NewStore creates a Store that is ready to be used.
func NewStore() *Store {
	s := &Store{
		m:        make(map[string]*Message),
		add:      make(chan *Message),
		request:  make(chan string),
		response: make(chan *Message),
	}
	go s.dispatch()
	return s
}

func (s *Store) dispatch() {
	for {
		select {
		case msg := <-s.add:
			s.m[msg.UUID] = msg
		case id := <-s.request:
			value, _ := s.m[id]
			s.response <- value
		}
	}
}

// Get retrieves the message with a UUID from the store.
func (s *Store) Get(uuid string) *Message {
	s.request <- uuid
	return <-s.response
}

// Add inserts the given message into the store.
func (s *Store) Add(msg *Message) {
	s.add <- msg
}
