# 6.824-fp
Final Project

In this repository, we have successfully implemented a command line interface text editor, such that multiple clients can collaborate on a single document and attain eventual consistency in the document through the use of operational transforms.

The paper writeup is available [here](6.824.pdf).

## How to run

~~~~
git clone https://github.com/qandrew/6.824-fp.git
cd 6.824-fp
export GOPATH=$PWD
export PATH=$PATH:$GOPATH/bin
cd $GOPATH/src/client
go install client
cd $GOPATH/src/server_test
go install server_test
~~~~
to run the client, open a terminal and run
~~~~
cd $GOPATH/bin
client
~~~~

to run the server, open a separate terminal and run
~~~~
cd $GOPATH/bin
server_test
~~~~

If errors occur while attempting to install `server_test` or `client`, be sure to reinstall the dependencies. Try:
~~~~
cd $GOPATH/src/github.com
rm -rf *
go get github.com/jroimartin/gocui
go get github.com/mattn/go-runewidth
go get github.com/nsf/termbox-go
~~~~
