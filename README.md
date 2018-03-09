# Arbor

Arbor is an experimental chat protocol that models a conversation
as a tree of messages instead of an ordered list. This means that
the conversation can organically diverge into several conversations
without the conversations appearing interleaved.

Arbor is unbelievably primitive right now. With time, it may develop
into something usable, but right now you can't even send messages on
the default client.

## Testing Arbor
If you'd like to see where things stand, you should be able to do the following:

1. Run `go build` in the root directory to create an `arbor` executable. This will listen on port `7777` for clients.
2. `cd client` and `go build` in there to create a `client` executable that you will use to connect.
3. `cd kudzu` and `go build` to create the `kudzu` test client. This just creates fake messages to simulate having real users.
4. First, run `./arbor`
5. In a different terminal, run `./kudzu localhost:7777`
6. In yet another terminal, run `./client localhost:7777 2> out`. Please note that you must redirect stderr to prevent the log from interfering with the UI.
7. Mess around in the client UI. Arrow keys and vim-like movements are supported. Ctrl-C will exit.


## Future Work

Things to do:
- ~~Make the client aware of the relationship between the view that holds a message and the message itself. This will allow for replying easily.~~
- ~~Build a better client-side data structure for modeling the conversation tree that would support fast lookup times for siblings.~~
- ~~Implement a test client that creates random subtrees as the server runs.~~
- ~~Implement a graphical indicator for when a message has siblings.~~
- ~~Implement scrolling through history relative to an arbitrary cursor, rather than the current leaf message~~
- ~~Implement subtree switching and find a good heuristic for choosing the default path within a subtree the first time that you view it~~
- Implement replies (easy once the other stuff is done).
- Implement a more robust protocol with version numbers, length indicators
  (for fast processing), usernames, and timestamps.
- Investigate arbor server clustering by having a new server connect as a client to an old one.
- Fix JSON parser so that all stacked messages are processed.
