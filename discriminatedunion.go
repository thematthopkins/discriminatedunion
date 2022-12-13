package discriminatedunion

import (
	"fmt"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"strings"
)

func (f *DiscriminatedUnionFact) AFact() {}
func (f *DiscriminatedUnionFact) String() string {
	return strings.Join(f.Types, ", ")
}

type DiscriminatedUnionFact struct {
	Types []string
}

var Analyzer = &analysis.Analyzer{
	Name:      "discriminated",
	Doc:       "check exhaustive pattern matching for discriminated unions",
	Run:       run,
	Requires:  []*analysis.Analyzer{inspect.Analyzer},
	FactTypes: []analysis.Fact{&DiscriminatedUnionFact{}},
}

func isDiscriminatedUnionSignifier(methodName string, typeName string) bool {
	return strings.HasSuffix(methodName, typeName) &&
		strings.HasPrefix(methodName, "Is") &&
		methodName == ("Is"+typeName)
}

func typeNameToSUnionSignifierMethodName(typeName string) string {
	return "Is" + typeName
}

type NameWithInterface struct {
	Name          *ast.Ident
	Interfaze     *ast.InterfaceType
	SignifierFunc *ast.FuncType
}

func typeToString(t types.Type) string {
	switch cased := t.(type) {
	case *types.Named:
		definingObj := cased.Obj()
		return fmt.Sprintf("%s.%s", definingObj.Pkg().Name(), definingObj.Name())
	case *types.Pointer:
		return "*" + typeToString(cased.Elem())
	default:
		panic(fmt.Sprintf("invalid type: %T", t))
	}
}

func signifierNameToInterface(inspector *inspector.Inspector) map[string]NameWithInterface {
	signifierNameToInterface := make(map[string]NameWithInterface)

	inspector.Preorder([]ast.Node{&ast.TypeSpec{}}, func(n ast.Node) {
		decl := n.(*ast.TypeSpec)
		interfaze, ok := decl.Type.(*ast.InterfaceType)
		if !ok {
			return
		}
		typeName := decl.Name.Name
		discriminatedUnionFunc := (*ast.FuncType)(nil)
		for _, method := range interfaze.Methods.List {
			if method.Names == nil {
				continue
			}
			methodName := method.Names[0].Name
			if isDiscriminatedUnionSignifier(methodName, typeName) {
				discriminatedUnionFunc = method.Type.(*ast.FuncType)
			}
		}
		if discriminatedUnionFunc == nil {
			return
		}

		signifierNameToInterface[typeNameToSUnionSignifierMethodName(typeName)] = NameWithInterface{
			Name:          decl.Name,
			Interfaze:     interfaze,
			SignifierFunc: discriminatedUnionFunc,
		}
	})

	return signifierNameToInterface
}

func interfaceToMembers(inspector *inspector.Inspector, typesInfo *types.Info, signifierNameToInterface map[string]NameWithInterface) map[NameWithInterface]([]types.Type) {
	interfaceToMembers := make(map[NameWithInterface]([]types.Type))

	inspector.Preorder([]ast.Node{&ast.FuncDecl{}}, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if fn.Recv == nil {
			return
		}

		if len(fn.Recv.List) != 1 {
			panic("unexpected receiver count")
		}

		matchingInterface, hasMatch := signifierNameToInterface[fn.Name.Name]
		if !hasMatch {
			return
		}

		receiverType := typesInfo.TypeOf(fn.Recv.List[0].Type)

		interfaceToMembers[matchingInterface] = append(interfaceToMembers[matchingInterface], receiverType)
	})

	return interfaceToMembers
}

func exportDiscriminatedUnionFacts(interfaceToMembers map[NameWithInterface]([]types.Type), pass *analysis.Pass) {
	for interfaze, members := range interfaceToMembers {
		memberTypeNames := make([]string, 0, len(members))
		for _, v := range members {
			memberTypeNames = append(memberTypeNames, typeToString(v))
		}
		pass.ExportObjectFact(
			pass.TypesInfo.Defs[interfaze.Name],
			&DiscriminatedUnionFact{
				Types: memberTypeNames,
			},
		)
	}
}

func createSwitchStatementWarnings(inspector *inspector.Inspector, pass *analysis.Pass) {
	inspector.Preorder([]ast.Node{&ast.TypeSwitchStmt{}}, func(n ast.Node) {
		statement := n.(*ast.TypeSwitchStmt)
		assignExpr, isAssign := statement.Assign.(*ast.AssignStmt)
		expr := (ast.Expr)(nil)
		if isAssign {
			if len(assignExpr.Lhs) > 1 {
				panic("lhs unsupported")
			}
			if len(assignExpr.Rhs) > 1 {
				panic("rhs unsupported")
			}
			expr = assignExpr.Rhs[0]
		} else {
			expr = statement.Assign.(*ast.ExprStmt).X
		}

		t := expr.(*ast.TypeAssertExpr)
		if t.Type != nil {
			panic("unexpected type assertion")
		}
		typeInfo := pass.TypesInfo.TypeOf(t.X)
		tagType, isNamed := typeInfo.(*types.Named)
		if !isNamed {
			return
		}

		typeFact := DiscriminatedUnionFact{}
		foundType := pass.ImportObjectFact(tagType.Obj(), &typeFact)
		if !foundType {
			return
		}

		coveredTypes := switchStatementTypes(pass, statement.Body)
		missingCases := ([]string)(nil)
		for _, v := range typeFact.Types {
			if _, hasMatch := coveredTypes[v]; !hasMatch {
				missingCases = append(missingCases, v)
			}
		}

		if len(missingCases) > 0 {
			missingCasesStr := strings.Join(missingCases, ", ")
			pass.Reportf(statement.End(), "missing cases for discriminated union types: %s", missingCasesStr)
		}
	})
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	signifierNameToInterface := signifierNameToInterface(inspector)
	interfaceToMembers := interfaceToMembers(inspector, pass.TypesInfo, signifierNameToInterface)
	exportDiscriminatedUnionFacts(interfaceToMembers, pass)
	createSwitchStatementWarnings(inspector, pass)
	return nil, nil
}

func switchStatementTypes(pass *analysis.Pass, body *ast.BlockStmt) map[string]struct{} {
	result := make(map[string]struct{}, len(body.List))
	for _, v := range body.List {
		caseClause := v.(*ast.CaseClause)
		for _, expr := range caseClause.List {
			typeInfo := pass.TypesInfo.TypeOf(expr)
			result[typeToString(typeInfo)] = struct{}{}
		}
	}
	return result
}
