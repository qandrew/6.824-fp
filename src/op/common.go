package op

import (
	"strings"
)

type Op struct {
	OpType   string // delete, insert, ??
	Position int    // location of where we are operationg
	Version  int    // version of the text that we are operating on
	VersionS int    // version of server
	Uid      int64  // uid of server
	Payload  string // in case it is an insert or move
	Path     string // path of the file
}

type Snapshot struct {
	Value    string
	Uid      int64 //
	Version  int   // version of client
	VersionS int   // version of server
}

type OpReply struct {
	// this structure contains all operation logs
	Logs []Op
	Num  int // idk debug purpose?
}

func ApplyOperation(args Op, currState string) string {
	// no OT needed, simply insert/delete args onto the currstate string
	if args.OpType == "ins" {
		// currState += args.Payload
		if args.Position == 0 {
			currState = args.Payload + currState // append at beginning
		} else {
			currState = currState[:args.Position] + args.Payload + currState[args.Position:]
		}
	} else if args.OpType == "del" {
		// if args.Position == 0{
		//   return // don't apply if we are at 0
		// }
		if args.Position == len(currState) && len(currState) != 0 {
			currState = currState[:args.Position-1]
		} else {
			currState = currState[:args.Position-1] + currState[args.Position:]
		}
	}
	// fmt.Println("applied", args, "now", currState)

	return currState
}

// Takes two Ops opC and opS and transforms them to opC' and opS'. They
// Must follow the property that performing opC followed by opS' results
// in the same state as opS followed by opC'. We use the convention that
// the first operation is client side and the second one is server side.
func Xform(opC Op, opS Op) (Op, Op) {
	opTypeC := opC.OpType
	opTypeS := opS.OpType

	newOpC := opC
	newOpS := opS

	if (opTypeC == "ins" || opTypeC == "del") && (opTypeS == "ins" || opTypeS == "del") && opC.Path != opS.Path {
		// two operations on unrelated files
		return opC, opS
	}

	if opTypeC == "ins" && opTypeS == "ins" {
		posC := newOpC.Position
		posS := newOpS.Position
		if posC > posS {
			newOpC.Position += len(opS.Payload)
		} else if posS > posC {
			newOpS.Position += len(opC.Payload)
		} else { // Tiebreaking rules: the one with the larger Uid comes later (ie gets its position modified)
			if newOpC.Uid > newOpC.Uid {
				newOpC.Position += len(opS.Payload)
			} else {
				newOpS.Position += len(opC.Payload)
			}
		}
		return newOpC, newOpS
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
		return newOpC, newOpS
	}

	if opTypeC == "del" && opTypeS == "ins" {
		posC := newOpC.Position
		posS := newOpS.Position
		if posC > posS {
			newOpC.Position += len(opS.Payload)
		} else {
			newOpS.Position--
		}
		return newOpC, newOpS
	}

	if opTypeC == "ins" && opTypeS == "del" {
		posC := newOpC.Position
		posS := newOpS.Position
		if posC < posS {
			newOpS.Position += len(opC.Payload)
		} else {
			newOpC.Position--
		}
		return newOpC, newOpS
	}

	if opTypeC == "move" && opTypeS == "move" {
		newOpC.Path = replacePrefix(opC.Path, opS.Path, opS.Payload)
		newOpS.Path = replacePrefix(opS.Path, opC.Path, opC.Payload)

		return newOpC, newOpS
	}

	if opTypeC == "move" {
		newOpS.Path = replacePrefix(opS.Path, opC.Path, opC.Payload)

		return newOpC, newOpS
	}

	if opTypeS == "move" {
		newOpC.Path = replacePrefix(opC.Path, opS.Path, opS.Payload)

		return newOpC, newOpS
	}

	return opC, opS
}

// works on paths, ex.:
//       replacePrefix("/foo/bar", "/foo", "/baz")
//       returns: "/baz/bar"
func replacePrefix(orig, prefix, newPrefix string) string {
	if orig == prefix {
		return newPrefix
	}
	if strings.HasPrefix(orig, prefix+"/") {
		return newPrefix + orig[len(prefix):]
	}
	return orig
}
