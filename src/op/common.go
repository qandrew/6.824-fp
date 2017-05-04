package op

type Op struct {
  OpType   string // delete, insert, ??
  Position int    // location of where we are operationg
  Version  int    // version of the text that we are operating on
  VersionS int    // version of server
  Uid      int64  // uid of server
  Payload  string // in case it is an insert
}

type Snapshot struct {
  Value 	string
  Uid 	int64	//
  Version int  	// version of client
  VersionS int 	// version of server
}

type OpReply struct {
  // this structure contains all operation logs
  Logs      []Op
  Num		int // idk debug purpose?
}

// Applies an operation to a string. Returns an error if the operation is not possible w/ the string
/*
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
*/

// Takes two Ops opC and opS and transforms them to opC' and opS'. They
// Must follow the property that performing opC followed by opS' results
// in the same state as opS followed by opC'. We use the convention that
// the first operation is client side and the second one is server side.
func xform(opC Op, opS Op) (Op, Op) {
  opTypeC := opC.OpType
  opTypeS := opS.OpType

  newOpC := opC
  newOpS := opS

  // If there's a lot of different types of operations, we probably
  // Need a matrix, but if it's just insert/delete1 this works for now.

  if opTypeC == "ins" && opTypeS == "ins" {
    posC := newOpC.Position
    posS := newOpS.Position
    if posC > posS {
      newOpC.Position += len(opS.Payload)
    } else if posS > posC {
      newOpS.Position += len(opC.Payload)
    } else { // They are equal, tiebreak with server comes first or something
	     // We may need an additional flag in the Op struct that's something
	     // like isFromServer
	     // :thinking:
    }
  }

  if opTypeC == "del" && opTypeS == "del" {
    posC := newOpC.Position
    posS := newOpS.Position
    if posC > posS {
      newOpC.Position--
    } else if posS > posC {
      newOpS.Position--
    } else {
      // In this case both the server and the client are trying to delete
      // the same thing. So in this case we have double no op.
      newOpC.OpType = "noOp"
      newOpS.OpType = "noOp"
    }
  }

  // One insert, one delete. Maybe some kind of matrix is more elegant.
  if (opTypeC == "del" && opTypeS == "ins") {
    posC := newOpC.Position
    posS := newOpS.Position
    if posC > posS {
      newOpC.Position += len(opS.Payload)
    } else {
      newOpS.Position--
    }
  }

  if (opTypeC == "ins" && opTypeS == "del") {
    posC := newOpC.Position
    posS := newOpS.Position
    if posC < posS {
      newOpS.Position += len(opC.Payload)
    } else {
      newOpC.Position--
    }
  }

  return newOpC, newOpS
}
