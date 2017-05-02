package op

type Op struct {
	OpType   string // delete, insert, ??
	Position int    // location of where we are operationg
	Version  int    // version of the text that we are operating on
	VersionS int    // version of server
	Uid		 int64 	// unique id of client
	Payload  string // in case it is an insert
}

type Snapshot struct {
	Value string
}

// type Logs struct {
//   // this structure contains all operation logs
//   Logs      []Op
// }
