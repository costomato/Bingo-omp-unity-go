## Server side code for Bingo game

### Written in Go

This part handles websockets and other backend logics. Basically, here we do the task of data transfer between player 1 and player 2.

Checking for the winner is not done here.

I have used Gorilla for upgrading normal http connection to websocket. Hence, we achieve full duplex communication.