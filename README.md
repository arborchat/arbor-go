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

```
go get github.com/whereswaldon/arbor/cmd/...
```

Then:
1. First, run `arbor`
2. In a different terminal, run `kudzu localhost:7777`
3. In yet another terminal, run `pergola localhost:7777 2> out`. Please note that you must redirect stderr to prevent the log from interfering with the UI.
4. Mess around in the client UI. Arrow keys are supported. Ctrl-C will exit.

## Controls

Up/Down - Move cursor forward/backward in current thread view
Left/Right - If the message under the cursor has siblings in the tree, switch to them and follow that history to a leaf message
Enter - Compose a reply to the highlighted message (press enter again to send)

## Future Work

Things to do:
- ~~Make the client aware of the relationship between the view that holds a message and the message itself. This will allow for replying easily.~~
- ~~Build a better client-side data structure for modeling the conversation tree that would support fast lookup times for siblings.~~
- ~~Implement a test client that creates random subtrees as the server runs.~~
- ~~Implement a graphical indicator for when a message has siblings.~~
- ~~Implement scrolling through history relative to an arbitrary cursor, rather than the current leaf message~~
- ~~Implement subtree switching and find a good heuristic for choosing the default path within a subtree the first time that you view it~~
- ~~Implement replies (easy once the other stuff is done).~~
- Implement a visual notification of unread messages
- ~~Implement a more robust protocol with version numbers, length indicators~~
  (for fast processing), usernames, and timestamps.
- Investigate arbor server clustering by having a new server connect as a client to an old one.
- Fix JSON parser so that all stacked messages are processed.
