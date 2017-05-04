package op

type Op struct {
  OpType   string // delete, insert, ??
  Position int    // location of where we are operationg
  Version  int    // version of the text that we are operating on
  VersionS int    // version of server
  Payload  string // in case it is an insert
  Uid      int64  // uid of server
}

type Snapshot struct {
  Value 	string
  Uid 	int64	//
  Version int  	// version of client
  VersionS int 	// version of server
}

// type Logs struct {
//   // this structure contains all operation logs
//   Logs      []Op
// }

// Applies an operation to a string. Returns an error if the operation is not possible w/ the string
func applyOp(op Op, text string) string {
  newText := text
  if op.OpType == "ins" {
    if op.Position <= len(string) {
      
    } else {
      fmt.PrintLn("Trying to insert after the end of string")
    }
  } else if op.OpType == "del" {

  }

}

// Takes two Ops op1 and op2 and transforms them to op1' and op2'. They
// Must follow the property that performing op1 followed by op2' results
// in the same state as op2 followed by op1'
func xform(op1 Op, op2 Op) (Op, Op) {
  opType1 := op1.OpType
  opType2 := op2.OpType

  newOp1 := op1
  newOp2 := op2

  // If there's a lot of different types of operations, we probably
  // Need a matrix, but if it's just insert/delete1 this works for now.

  if opType1 == "ins" && opType2 == "ins" {
    pos1 := newOp1.Position
    pos2 := newOp2.Position
    if pos1 > pos2 {
      newOp1.Position += len(op2.Payload)
    } else if pos2 > pos1 {
      newOp2.Position += len(op1.Payload)
    } else { // They are equal, tiebreak with server comes first or something
	     // We may need an additional flag in the Op struct that's something
	     // like isFromServer
      // :thinking:
    }
  }

  if opType1 == "del" && opType2 == "del" {
    pos1 := newOp1.Position
    pos2 := newOp2.Position
    if pos1 > pos2 {
      newOp1.Position--
    } else if pos2 > pos1 {
      newOp2.Position--
    } else {
      // In this case both the server and the client are trying to delete
      // the same thing. So in this case we have double no op.
      newOp1.OpType = "noOp"
      newOp2.OpType = "noOp"
    }
  }

  // One insert, one delete. Maybe some kind of matrix is more elegant.
  if (opType1 == "del" && opType2 == "ins") {
    pos1 := newOp1.Position
    pos2 := newOp2.Position
    if pos1 > pos2 {
      newOp1.Position += len(op2.Payload)
    } else {
      newOp2.Position--
  }

  if (opType1 == "ins" && opType2 == "del") {
    pos1 := newOp1.Position
    pos2 := newOp2.Position
    if pos1 < pos2 {
      newOp2.Position += len(op1.Payload)
    } else {
      newOp1.Position--
  }

  return newOp1, newOp2
}
