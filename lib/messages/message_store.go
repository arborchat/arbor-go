package messages

type Store struct {
	m        map[string]*Message
	add      chan *Message
	request  chan string
	response chan *Message
}

func NewStore() *Store {
	s := &Store{
		m: make(map[string]*Message),
		add: make(chan *Message),
		request: make(chan string),
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

func (s *Store) Get(uuid string) *Message {
	s.request <- uuid
	return <-s.response
}

func (s *Store) Add(msg *Message) {
	s.add <- msg
}
