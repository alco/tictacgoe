A simple peer-to-peer game between two players.

## Program spec

```
Usage: ttt [-a ADDRESS] [-p PORT]

  -a ADDRESS  Start the program in client-mode and try to connect to the remote address

  -p PORT     Use the specified port when connecting or listening for new connections.
              (default: 8888)

```

When the `-a` flag is omitted, the program listens on port `PORT` and also accepts user input. Typing `connect <address>:<port>` will try to connect to the given address and a start a game.

As soon as connection has been established, the game starts.

The first player is selected randomly. Each player makes a move by typing in the coordinates. The program prints board state between the turns for each player. While one player is deciding on their move, the other one has to wait for their turn.

### Example session

```
$ ttt

Welcome to Tic-Tac-Toe. Listening on port 8888.

> connect 127.0.0.1
Trying to connect...

*** Game started ***

You play X.

   1   2   3
a [ ] [ ] [ ]
b [ ] [ ] [ ]
c [ ] [ ] [ ]

Your turn.
> a 1

   1   2   3
a [X] [ ] [ ]
b [ ] [ ] [ ]
c [ ] [ ] [ ]

Waiting for opponent...

   1   2   3
a [X] [ ] [ ]
b [ ] [O] [ ]
c [ ] [ ] [ ]

Your turn.
>

...

*** You win! ***

- OR -

*** You lose! ***

- OR -

*** It's a draw ***
```
