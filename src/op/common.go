package op

type Op struct {
	OpType   string // del, ins, empty, good
	Position int    // location of where we are operationg
	Version  int    // version of the text that we are operating on
	VersionS int    // version of server
	Uid		 int64  // uid of server
	Payload  string // in case it is an insert
}

type Snapshot struct {
	Value 	string
	// Uid 	int64	//
	Version int  	// version of client
	VersionS int 	// version of server
}

// type Logs struct {
//   // this structure contains all operation logs
//   Logs      []Op
// }
