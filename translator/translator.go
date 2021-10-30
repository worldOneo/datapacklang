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

	ifScore = "execute %s score %s %s %s %s %s run "
	as      = "execute as %s run "
	result  = "execute store result %s %s run %s "
)

const (
	storeIf  = "if"
	storeNot = "unless"
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
		return []command{fmt.Sprintf(createStorage, T.getStore(n.Identifier))}, nil
	case ast.If:
		return T._if(n)
	case ast.String:
		return []command{n.Value}, nil
	case ast.As:
		prefix := fmt.Sprintf(as, n.Selector)
		cmds, err := T.Translate(n.Body)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(cmds); i++ {
			cmds[i] = prefix + cmds[i]
		}
		return cmds, nil
	}
	return []command{}, nil
}

func (T *Translator) _if(n ast.If) ([]command, error) {
	cmds := make([]command, 0)

	operator := storeIf
	if n.Not {
		operator = storeNot
	}

	a := T.registers.claim(T)
	b := T.registers.claim(T)

	// Short if optimization
	var prefix string
	aAc, okFirst := n.First.(ast.StoreAccess)
	bAc, okSecond := n.First.(ast.StoreAccess)
	if len(n.Body.Body) == 1 && okFirst && okSecond {
		aV := T.trueName(aAc.Identifier)
		bV := T.trueName(bAc.Identifier)
		prefix = fmt.Sprintf(ifScore, operator, aV, aAc.Store, conditionalOperators[n.Comparator], bV, bAc.Store)
	} else {
		ok := T.createStore(dplTemp)
		if !ok {
			cmd, err := T.Translate(ast.CreateStore{dplTemp})
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, cmd...)
		}

		temp := T.getStore(dplTemp)
		leftRegister := ast.MakeStoreAssign(dplTemp, a, true, tokens.OperationSet, n.First)
		rightRegister := ast.MakeStoreAssign(dplTemp, b, true, tokens.OperationSet, n.Second)
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

		prefix = fmt.Sprintf(ifScore, operator, aV, temp, conditionalOperators[n.Comparator], bV, temp)
	}
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
	ok := T.createStore(dplTemp)
	if !ok {
		cmd, err := T.Translate(ast.CreateStore{dplTemp})
		if err != nil {
			return nil, ast.StoreAccess{}, err
		}
		cmds = append(cmds, cmd...)
	}
	a := T.registers.claim(T)
	b := T.registers.claim(T)
	initRegister := ast.MakeStoreAssign(dplTemp, a, true, tokens.OperationSet, n.First)
	calcRegister := ast.MakeStoreAssign(dplTemp, b, true, tokens.OperationSet, n.Second)
	operations := ast.MakeStoreAssign(dplTemp, a, true, n.Operator, ast.MakeStoreAccess(dplTemp, b, true))
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
	return cmds, ast.MakeStoreAccess(dplTemp, a, true), nil
}

func (T *Translator) storeAssign(n ast.StoreAssign) ([]command, error) {
	store := T.getStore(n.Store)
	variable := T.getVariable(n.Identifier.Identifier)
	if !n.Identifier.IsVar {
		variable = n.Identifier.Identifier
	}
	value := n.Value
	switch v := value.(type) {
	case ast.Int:
		op, ok := storageAssignOperations[n.Operation]
		if !ok {
			return nil, fmt.Errorf("Invalid operator")
		}
		return []command{fmt.Sprintf(editStorage, op, variable, store, v.Value)}, nil
	case ast.StoreAccess:
		op, ok := storageAccessOperations[n.Operation]
		if !ok {
			return nil, fmt.Errorf("Invalid operator")
		}
		withStore := T.getStore(v.Store)
		withVar := T.getVariable(v.Identifier.Identifier)
		if !n.Identifier.IsVar {
			withVar = n.Identifier.Identifier
		}
		return []command{fmt.Sprintf(storageOperation, variable, store, op, withVar, withStore)}, nil
	case ast.Calculation:
		cmds, access, err := T.resolveCalculation(v)
		if err != nil {
			return nil, err
		}
		assign, err := T.storeAssign(ast.MakeStoreAssign(n.Store, n.Identifier.Identifier, n.Identifier.IsVar, n.Operation, access))
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, assign...)
		return cmds, nil
	case ast.String:
		return []command{fmt.Sprintf(result, variable, store, v.Value)}, nil
	}
	return nil, fmt.Errorf("Invalid assignment")
}

func (T *Translator) getStore(key string) string {
	_, ok := T.stores[key]
	if !ok {
		T.stores[key] = T.nextIdentifier()
	}
	return T.stores[key]
}

func (T *Translator) getVariable(variable string) string {
	v, ok := T.variables[variable]
	if !ok {
		v = T.nextIdentifier()
		T.variables[variable] = v
	}
	return v
}

func (T *Translator) createStore(variable string) bool {
	_, ok := T.stores[variable]
	if !ok {
		T.stores[variable] = T.nextIdentifier()
	}
	return ok
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

func (T *Translator) trueName(index ast.Index) string {
	if index.IsVar {
		return T.getVariable(index.Identifier)
	}
	return index.Identifier
}
