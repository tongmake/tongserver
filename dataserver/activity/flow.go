package activity

import (
	"fmt"
	"tongserver.dataserver/utils"
)

const (
	F_TO     int = 0
	F_IFTO   int = 1
	F_IFLOOP int = 2
)

type FlowResult uint8

const (
	FR_ERROR   FlowResult = 0
	FR_CONINUE FlowResult = 1
	FR_BREAK   FlowResult = 2
)

type IFlow interface {
	DoFlow(flowcontext IContext) (FlowResult, error)
}

type Flows struct {
	flows []IFlow
}

func (c *Flows) ExecuteFlows(flowcontext IContext) error {
	for _, item := range c.flows {
		rs, err := item.DoFlow(flowcontext)
		if err != nil {
			return err
		}
		if rs == FR_BREAK {
			break
		}
	}
	return nil
}

// 基本流程
type Flow struct {
	gate         int
	define       map[string]interface{}
	flowInstance FlowInstance
}

// 直通flow
type FlowTo struct {
	Flow
	activitys []*IActivity
}

// 条件flow
type FlowIfTo struct {
	Flow
	expression  string
	thenAct     []*IActivity
	elseThenAct []*IActivity
}

// 循环flow
type FlowLoop struct {
	Flow
	assignExpression []string
	whileExpression  string
	stepExcpression  []string
	dotarget         []*IActivity
}

func (c *Flow) executeActivitys(activitys []*IActivity, flowcontext IContext) (FlowResult, error) {
	for _, v := range activitys {
		err := (*v).Execute(flowcontext)
		if err != nil {
			return FR_BREAK, err
		}
	}
	return FR_CONINUE, nil
}

func (c *FlowTo) DoFlow(flowcontext IContext) (FlowResult, error) {
	return c.executeActivitys(c.activitys, flowcontext)
}

func (c *FlowIfTo) DoFlow(flowcontext IContext) (FlowResult, error) {
	r, err := DoExpressionBool(c.expression, flowcontext)
	if err != nil {
		return FR_ERROR, fmt.Errorf("执行表达式 %s 发生错误,%s", c.expression, err.Error())
	}
	if r {
		return c.executeActivitys(c.thenAct, flowcontext)
	} else {
		if c.elseThenAct != nil {
			return c.executeActivitys(c.elseThenAct, flowcontext)
		}
	}
	return FR_CONINUE, nil

}

func (c *FlowLoop) DoFlow(flowcontext IContext) (FlowResult, error) {
	if len(c.assignExpression) != 0 {
		err := ExecuteExpressions(flowcontext, c.assignExpression)
		if err != nil {
			return FR_ERROR, err
		}
	}
	for {
		b, err := DoExpressionBool(c.whileExpression, flowcontext)
		if err != nil {
			return FR_ERROR, err
		}
		if !b {
			break
		}
		ft, err := c.executeActivitys(c.dotarget, flowcontext)
		if err != nil {
			return FR_ERROR, err
		}
		if ft == FR_BREAK {
			break
		}
		if len(c.stepExcpression) != 0 {
			err := ExecuteExpressions(flowcontext, c.stepExcpression)
			if err != nil {
				return FR_ERROR, err
			}
		}
	}
	return FR_CONINUE, nil
}

// {
//        "gate": "to",
//        "activity1": {
//			"expressions":["var_b=var_b+10"],
//			"style":"stdout"
//		}
//	}
func NewFlowTo(define map[string]interface{}, flowInstance *FlowInstance) (IFlow, error) {
	target := utils.GetArrayFromMap(define, "target")
	if target == nil {
		return nil, fmt.Errorf("创建toflow失败,没有target属性")
	}
	acts, err := CreateActivitys(target, flowInstance)
	if err != nil {
		return nil, err
	}
	f := &FlowTo{
		Flow: Flow{
			gate: F_TO,
		},
		activitys: acts,
	}
	return f, nil
}

//{
//    "gate":"ifto",
//    "if":"表达式",
//    "then":{},
//    "else":{}
//}
func NewFlowIfTo(define map[string]interface{}, flowInstance *FlowInstance) (IFlow, error) {
	f := &FlowIfTo{
		Flow: Flow{
			gate: F_IFTO,
		},
	}
	exp, ok := define["if"]
	if !ok {
		return nil, fmt.Errorf("创建ifto失败，没有if属性")
	}
	f.expression = exp.(string)
	then := utils.GetArrayFromMap(define, "then")
	if then == nil {
		return nil, fmt.Errorf("创建ifto失败，没有then属性")
	}
	am, err := CreateActivitys(then, flowInstance)
	if err != nil {
		return nil, fmt.Errorf("创建ifto失败，then创建失败，%s", err.Error())
	}
	f.thenAct = am
	els := utils.GetArrayFromMap(define, "else")
	if els != nil {
		f.elseThenAct, err = CreateActivitys(els, flowInstance)
		if err != nil {
			return nil, fmt.Errorf("创建ifto失败，else 创建失败，%s", err.Error())
		}
	}
	return f, nil
}

// 			{
//                "gate":"loop",
//                "assign":["",""],
//                "while":"",
//                "step":["",""],
//                "do":{
//                    "activity2":{}
//                }
//          }
// 创建循环flow
func NewFlowLoop(def map[string]interface{}, flowInstance *FlowInstance) (IFlow, error) {
	f := &FlowLoop{
		Flow: Flow{
			gate: F_IFLOOP,
		},
	}
	m := def["assign"]
	if m != nil {
		dd := m.([]interface{})
		f.assignExpression = make([]string, len(dd), len(dd))
		for index, d := range dd {
			f.assignExpression[index] = d.(string)
		}
	}
	m = def["step"]
	if m != nil {
		dd := m.([]interface{})
		f.stepExcpression = make([]string, len(dd), len(dd))
		for index, d := range dd {
			f.stepExcpression[index] = d.(string)
		}
	}
	wh, ok := def["while"]
	if !ok {
		return nil, fmt.Errorf("创建flowLoop失败，while属性是必须的")
	}
	f.whileExpression = wh.(string)
	doacts := utils.GetArrayFromMap(def, "do")
	sct, err := CreateActivitys(doacts, flowInstance)
	if err != nil {
		return nil, fmt.Errorf("创建flowLoop失败，创建do节点失败，%s", err.Error())
	}
	f.dotarget = sct
	return f, nil
}

func NewFlow(d map[string]interface{}, flowInstance *FlowInstance) (IFlow, error) {
	gate, ok := d["gate"]
	if !ok {
		return nil, fmt.Errorf("缺少gate属性")
	}
	sg, ok := gate.(string)
	if !ok {
		return nil, fmt.Errorf("gate属性类型必须是string")
	}
	f, ok := flowCreatorFunContainer[sg]
	if !ok {
		return nil, fmt.Errorf("没有找到gate属性为%s的构造器", sg)
	}
	return f(d, flowInstance)
}
func CreateFlows(flows []interface{}, inst *FlowInstance) ([]IFlow, error) {
	if len(flows) > 0 {
		iflows := make([]IFlow, len(flows), len(flows))
		for index, item := range flows {
			inf := utils.ConvertObj2Map(item)
			if inf == nil {
				return nil, fmt.Errorf("创建流程实例失败，start节点的flow属性必须为一个对象数组")
			}
			f, err := NewFlow(inf, inst)
			if err != nil {
				return nil, err
			}
			iflows[index] = f
		}
		return iflows, nil
	}
	return nil, nil
}

// 用于创建flow的构造器
type FlowCreatorFun func(define map[string]interface{}, flowInstance *FlowInstance) (IFlow, error)

// flow的构造器的容器
var flowCreatorFunContainer = make(map[string]FlowCreatorFun)

func RegisterFlowCreator(gateName string, f FlowCreatorFun) {
	flowCreatorFunContainer[gateName] = f
}
