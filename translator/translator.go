package translator

import (
	"fmt"

	"github.com/worldOneo/datapacklang/ast"
	"github.com/worldOneo/datapacklang/tokens"
)

const (
	createStorage    = "scoreboard objectives add %s dummy"
	editStorage      = "scoreboard players %s %s %s %d"
	storageOperation = "scoreboard players operation %s %s %s %s %s"

	ifValue = "execute if score %s %s matches %d run "
	ifRange = "execute if score %s %s matches %s run "
	ifScore = "execute if score %s %s %s %s %s run "
)

const (
	storeAdd = "add"
	storeSet = "set"
	storeSub = "remove"
)

const (
	dplInternal = "_dpl_internal"
	dplTemp     = "_dpl_tmp"
)

var storageAccessOperations = make(map[tokens.OperationType]string)
var storageAssignOperations = make(map[tokens.OperationType]string)
var conditionalOperators = make(map[tokens.OperationType]string)

func init() {
	storageAccessOperations[tokens.OperationAdd] = "+="
	storageAccessOperations[tokens.OperationSub] = "-="
	storageAccessOperations[tokens.OperationSet] = "="
	storageAccessOperations[tokens.OperationMod] = "%="
	storageAccessOperations[tokens.OperationMul] = "*="
	storageAccessOperations[tokens.OperationDiv] = "/="

	storageAssignOperations[tokens.OperationAdd] = storeAdd
	storageAssignOperations[tokens.OperationDec] = storeSub
	storageAssignOperations[tokens.OperationSet] = storeSet

	conditionalOperators[tokens.OperationEq] = "="
	conditionalOperators[tokens.OperationNeq] = "!="
	conditionalOperators[tokens.OperationGt] = ">"
	conditionalOperators[tokens.OperationGte] = ">="
	conditionalOperators[tokens.OperationLt] = "<"
	conditionalOperators[tokens.OperationLte] = "<="

}

type command = string

type Translator struct {
	variables map[string]string
	stores    map[string]string
	registers *Registers
	nextVar   int
}

func New() Translator {
	return Translator{
		make(map[string]string),
		make(map[string]string),
		NewRegisters(),
		-1,
	}
}

func (T *Translator) Translate(program ast.Node) ([]command, error) {
	switch n := program.(type) {
	case ast.Block:
		body := n.Body
		instructions := make([]command, 0)
		for _, node := range body {
			inst, err := T.Translate(node)
			if err != nil {
				return []command{}, err
			}
			instructions = append(instructions, inst...)
		}
		return instructions, nil
	case ast.StoreAssign:
		return T.storeAssign(n)
	case ast.CreateStore:
		T.createStore(n.Identifier)
		v, _ := T.getStore(n.Identifier)
		return []command{fmt.Sprintf(createStorage, v)}, nil
	case ast.If:
		return T._if(n)
	case ast.String:
		return []command{n.Value}, nil
	}
	return []command{}, nil
}

func (T *Translator) _if(n ast.If) ([]command, error) {
	cmds := make([]command, 0)
	_, ok := T.getStore(dplTemp)
	if !ok {
		cmd, err := T.Translate(ast.CreateStore{dplTemp})
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, cmd...)
	}
	temp, _ := T.getStore(dplTemp)
	a := T.registers.claim(T)
	b := T.registers.claim(T)
	leftRegister := ast.StoreAssign{a, dplTemp, tokens.OperationSet, n.First}
	rightRegister := ast.StoreAssign{b, dplTemp, tokens.OperationSet, n.Second}
	leftEval, err := T.Translate(leftRegister)
	if err != nil {
		return nil, err
	}
	rightEval, err := T.Translate(rightRegister)
	if err != nil {
		return nil, err
	}
	cmds = append(cmds, leftEval...)
	cmds = append(cmds, rightEval...)
	aV := T.getVariable(a)
	bV := T.getVariable(b)
	prefix := fmt.Sprintf(ifScore, aV, temp, conditionalOperators[n.Comparator], bV, temp)
	for _, elem := range n.Body.Body {
		commands, err := T.Translate(elem)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(commands); i++ {
			commands[i] = prefix + commands[i]
		}
		cmds = append(cmds, commands...)
	}
	T.registers.free(a)
	T.registers.free(b)
	return cmds, nil
}

func (T *Translator) resolveCalculation(n ast.Calculation) ([]command, ast.StoreAccess, error) {
	cmds := make([]command, 0)
	_, ok := T.getStore(dplTemp)
	if !ok {
		cmd, err := T.Translate(ast.CreateStore{dplTemp})
		if err != nil {
			return nil, ast.StoreAccess{}, err
		}
		cmds = append(cmds, cmd...)
	}
	a := T.registers.claim(T)
	b := T.registers.claim(T)
	initRegister := ast.StoreAssign{a, dplTemp, tokens.OperationSet, n.First}
	calcRegister := ast.StoreAssign{b, dplTemp, tokens.OperationSet, n.Second}
	operations := ast.StoreAssign{a, dplTemp, n.Operator, ast.StoreAccess{b, dplTemp}}
	init, err := T.Translate(initRegister)
	if err != nil {
		return nil, ast.StoreAccess{}, err
	}
	calc, err := T.Translate(calcRegister)
	if err != nil {
		return nil, ast.StoreAccess{}, err
	}
	op, err := T.Translate(operations)
	if err != nil {
		return nil, ast.StoreAccess{}, err
	}
	cmds = append(cmds, init...)
	cmds = append(cmds, calc...)
	cmds = append(cmds, op...)
	T.registers.free(a)
	T.registers.free(b)
	return cmds, ast.StoreAccess{a, dplTemp}, nil
}

func (T *Translator) storeAssign(n ast.StoreAssign) ([]command, error) {
	store, ok := T.getStore(n.Store)
	if !ok {
		return nil, fmt.Errorf("Store %s doesnt exist", n.Store)
	}
	variable := T.getVariable(n.Identifier)
	value, ok := n.Value.(ast.Int)
	if !ok {
		value, ok := n.Value.(ast.StoreAccess)
		if !ok {
			value, ok := n.Value.(ast.Calculation)
			if !ok {
				return []command{}, nil
			}
			cmds, access, err := T.resolveCalculation(value)
			if err != nil {
				return nil, err
			}
			assign, err := T.storeAssign(ast.StoreAssign{n.Identifier, n.Store, n.Operation, access})
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, assign...)
			return cmds, nil
		}
		op, ok := storageAccessOperations[n.Operation]
		if !ok {
			return nil, fmt.Errorf("Invalid operator")
		}
		withStore, ok := T.getStore(value.Store)
		if !ok {
			return nil, fmt.Errorf("Store %s doesnt exist", n.Store)
		}
		withVar := T.getVariable(value.Identifier)
		return []command{fmt.Sprintf(storageOperation, variable, store, op, withVar, withStore)}, nil
	}
	op, ok := storageAssignOperations[n.Operation]
	if !ok {
		return nil, fmt.Errorf("Invalid operator")
	}
	return []command{fmt.Sprintf(editStorage, op, variable, store, value.Value)}, nil
}

func (T *Translator) getStore(key string) (string, bool) {
	k, ok := T.stores[key]
	return k, ok
}

func (T *Translator) getVariable(variable string) string {
	v, ok := T.variables[variable]
	if !ok {
		v = T.nextIdentifier()
		T.variables[variable] = v
	}
	return v
}

func (T *Translator) createStore(variable string) {
	_, ok := T.stores[variable]
	if !ok {
		T.stores[variable] = T.nextIdentifier()
	}
}

func toString(i int) string {
	if i < 0 {
		return ""
	}
	return toString((i/26)-1) + string('a'+(rune(i)%26))
}

func (T *Translator) nextIdentifier() string {
	T.nextVar++
	return toString(T.nextVar)
}
