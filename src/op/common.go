package op

type Op struct {
  OpType    string // delete, insert, ??
  Position  int // location of where we are operationg
  Version   int // version of the text that we are operating on
}

// type Logs struct {
//   // this structure contains all operation logs
//   Logs      []Op
// }