# Arbor

Arbor is an experimental chat protocol that models a conversation
as a tree of messages instead of an ordered list. This means that
the conversation can organically diverge into several conversations
without the conversations appearing interleaved.

Arbor is unbelievably primitive right now. With time, it may develop
into something usable, but right now you can't even send messages on
the default client.

Things to do:
- Make the client aware of the relationship between the view that holds
  a message and the message itself. This will allow for replying easily.
- Build a better client-side data structure for modeling the conversation
  tree that would support fast lookup times for siblings.
- Implement a test client that creates random subtrees as the server runs.
- Implement a graphical indicator for when a message has siblings.
- Implement subtree switching and find a good heuristic for choosing the
  default path within a subtree the first time that you view it
- Implement replies (easy once the other stuff is done).
- Implement a more robust protocol with version numbers, length indicators
  (for fast processing), usernames, and timestamps.
