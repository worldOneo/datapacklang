package translator

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"

	"github.com/worldOneo/datapacklang/ast"
	"github.com/worldOneo/datapacklang/tokens"
)

const (
	createStorage    = "scoreboard objectives add %s dummy"
	editStorage      = "scoreboard players %s %s %s %d"
	storageOperation = "scoreboard players operation %s %s %s %s %s"
	ifValue          = "if score %s %s %s %s"
	ifRange          = "if score %s %s matches %s"
	ifScore          = "if score %s %s %s %s %s"
)

const (
	storeAdd = "add"
	storeSet = "set"
	storeSub = "remove"
)

var storageAccessOperations = make(map[tokens.OperationType]string)
var storageAssignOperations = make(map[tokens.OperationType]string)

func init() {
	storageAccessOperations[tokens.OperationAdd] = "+="
	storageAccessOperations[tokens.OperationDec] = "+="
	storageAccessOperations[tokens.OperationSet] = "="

	storageAssignOperations[tokens.OperationAdd] = storeAdd
	storageAssignOperations[tokens.OperationDec] = storeSub
	storageAssignOperations[tokens.OperationSet] = storeSet
}

type command = string

type Translator struct {
	variables map[string]string
	stores    map[string]string
	nextVar   *big.Int
}

func New() Translator {
	return Translator{
		make(map[string]string),
		make(map[string]string),
		big.NewInt(0),
	}
}

func (T Translator) Translate(program ast.Node) ([]command, error) {
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
		store, ok := T.getStore(n.Store)
		if !ok {
			return nil, fmt.Errorf("Store %s doesnt exist", n.Store)
		}
		variable := T.getVariable(n.Identifier)
		value, ok := n.Value.(ast.Int)
		if !ok {
			value, ok := n.Value.(ast.StoreAccess)
			if !ok {
				break
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
	case ast.CreateStore:
		T.createStore(n.Identifier)
		v, _ := T.getStore(n.Identifier)
		return []command{fmt.Sprintf(createStorage, v)}, nil
	}
	return []command{}, nil
}

func (T Translator) getStore(key string) (string, bool) {
	k, ok := T.stores[key]
	return k, ok
}

func (T Translator) getVariable(variable string) string {
	v, ok := T.variables[variable]
	if !ok {
		v = T.nextIdentifier()
		T.variables[variable] = v
	}
	return v
}

func (T Translator) createStore(variable string) {
	_, ok := T.stores[variable]
	if !ok {
		T.stores[variable] = T.nextIdentifier()
	}
}

func (T Translator) nextIdentifier() string {
	new := T.nextVar.Add(T.nextVar, big.NewInt(1))
	return strings.Replace(base64.StdEncoding.EncodeToString(new.Bytes()), "=", "", -1)
}
