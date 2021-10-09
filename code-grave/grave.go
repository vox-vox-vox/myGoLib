package main

// 记录自己写过的代码
// 删了怪可惜的，但是留着确实也毫无作用
// 放在这里，也是对已经逝去的时间的的最好交代吧
// 愿他们死得其所

// 1. 基于history的流程图匹配重构版
// 死亡原因：基于history接口，所以速度并没有快多少，准确率也没有提高，所以没什么优势

/**

func (s *WorkflowService) GetFlowChart2(wfID, runID string, domain string) (resp *types.FlowChartResponse, err error) {
	var taskInfo TaskInfo
	getIndex := make(map[string]int)
	getFather := make(map[int]int)
	getDoWhile := make(map[string]int)
	visitCount := make(map[int]int)
	edges := make(map[int]map[int]int)
	//nextSimpleNodes:=make(map[int]*map[string]int)
	if runID == "" {
		runID, _ = mojito.GCadenceClientsWithDomain[workflow.CadenceDomain(domain)].GetRunIDByWorkflowID(wfID)
	}
	req := &workflow.DescribeWorkflowReq{
		WorkflowID: wfID,
		RunID:      runID,
		Domain:     domain,
	}
	//根据workflowID和RunID查出historyEvents和excutionInfo
	historyEvents, executionInfo, err := s.GetDescription(req)
	if err != nil {
		return
	}

	//加缓存目前会出现版本不匹配的bug，考虑加版本号或直接用Query解决，目前先不加缓存
	//// 通过go-cache查出workflow-Meta，不存在就调用下面的GetDefinition
	//var flowMeta workflow.WorkflowMeta
	//wfNameSlice:= []rune(wfID)
	//wfNameSlice=wfNameSlice[:len(wfNameSlice)-37]//去掉uuid
	//wfName:=string(wfNameSlice)
	//flowMetaInterface,ok := WorkflowMetaCache.Get(wfName)
	//if !ok{
	//	flowMeta, err = cadence.GetDefinition(historyEvents)
	//	if err != nil {
	//		return
	//	}
	//	WorkflowMetaCache.Set(wfName,flowMeta,cache.DefaultExpiration)
	//}else{
	//	flowMeta,ok = flowMetaInterface.(workflow.WorkflowMeta)
	//	if !ok {
	//		err = fmt.Errorf("interface convert error")
	//		return
	//	}
	//}

	flowMeta, err := cadence.GetDefinition(historyEvents)
	if err != nil {
		return
	}

	resp, err = parseWorkflowExecution(executionInfo)
	if err != nil {
		return
	}
	if len(historyEvents) > 0 {
		resp.Input = string(historyEvents[0].WorkflowExecutionStartedEventAttributes.Input)
	}
	// 建图
	AddTemplateNode(&resp.FlowNode, &edges, &workflow.WorkflowTaskMeta{Name: "BEGIN", Type: stateChange})
	GenTemplates(&resp.FlowNode, &edges, flowMeta.Tasks...)
	AddSingleNode(&resp.FlowNode, &workflow.WorkflowTaskMeta{Name: "END", Type: stateChange})
	// 建必要的数据结构
	genNeedDataStructure(&resp.FlowNode, &edges, &getIndex, &getFather, &visitCount, &getDoWhile)
	// 解析historyEvents
	scheduleEvents, _, err := s.GetTaskInfo2(historyEvents, &taskInfo, getIndex, wfID)
	taskInfo.taskIdMap = append(taskInfo.taskIdMap, map[int64]workflow.TaskID{
		int64(len(historyEvents)): {
			WorkflowID: wfID,
			RunID:      executionInfo.WorkflowExecutionInfo.Execution.GetRunId(),
		},
	})
	if err != nil {
		return
	}

	//用于running和queuing状态的更新
	for _, pendingActivity := range executionInfo.PendingActivities {
		if taskInfo.taskStatus[*pendingActivity.ActivityID] != types.FlowNodeCompleteStatus && taskInfo.taskStatus[*pendingActivity.ActivityID] != types.FlowNodeSkipStatus {
			if *pendingActivity.State == shared.PendingActivityStateScheduled {
				taskInfo.taskStatus[*pendingActivity.ActivityID] = types.FlowNodeScheduleStatus
			} else if *pendingActivity.State == shared.PendingActivityStateStarted {
				taskInfo.taskStatus[*pendingActivity.ActivityID] = types.FlowNodeStartStatus
				taskInfo.taskAttempt[*pendingActivity.ActivityID] = strconv.Itoa(int(*(pendingActivity.Attempt)))
				taskInfo.taskStartTime[*pendingActivity.ActivityID] = *pendingActivity.LastStartedTimestamp
			}
		}
	}

	//getSimpleNodeIndex2(&resp.FlowNode,&edges,&nextSimpleNodes)
	matchFormHistory(&resp.FlowNode, taskInfo, &scheduleEvents, &getIndex, &getFather, &getDoWhile, wfID, runID)
	DumpEdges2(edges, &resp.FlowEdges)
	edgeOutputMap := make(map[string]*types.FlowEdge)
	for _, value := range resp.FlowEdges {
		edgeOutputMap[value.Source+"to"+value.Target] = value //装入map
	}
	lastScheduledTask := *scheduleEvents[len(scheduleEvents)-1].ActivityTaskScheduledEventAttributes.ActivityId
	linkAllEdges(&resp.FlowNode, &edges, &edgeOutputMap, &visitCount, &getDoWhile, lastScheduledTask, &getFather)
	edgeOutput := make([]*types.FlowEdge, 0)
	for _, v := range edgeOutputMap {
		edgeOutput = append(edgeOutput, &types.FlowEdge{
			Target: v.Target,
			Source: v.Source,
			Used:   v.Used,
		})
	}
	resp.FlowEdges = edgeOutput
	return
}


// getIndex: 根据refName迅速查出对应index的map
// getFather: 根据index迅速查出父节点index的map，若有多个父节点只取一个不影响结果
// visitCount: 记录index处节点有多少个父节点的map，用于剪枝优化
// getDoWhile: 记录各个节点所属的DoWhile节点的map，通过栈建立
func genNeedDataStructure(nodes *[]*types.FlowNode, edges *map[int]map[int]int, getIndex *map[string]int, getFather *map[int]int, visitCount *map[int]int, getDoWhile *map[string]int) {
	stack := list.New()
	stack.PushBack(0)
	for index, node := range *nodes {
		if node.DataType == "DO_WHILE" {
			stack.PushBack(index)
			(*visitCount)[index]--
		}
		if node.DataType == "LOOP_END" {
			stack.Remove(stack.Back())
		}
		(*getDoWhile)[node.RefName] = stack.Back().Value.(int)
		(*getIndex)[node.RefName] = index
		for k, _ := range (*edges)[index] {
			(*getFather)[k] = index
			(*visitCount)[k]++
		}
	}
	return
}


func (s *WorkflowService) GetTaskInfo2(historyEvents []*shared.HistoryEvent, taskInfo *TaskInfo, flowNodes map[string]int, wfID string) (scheduleEvents []*shared.HistoryEvent, childEvents []map[string]map[string]string, err error) {
	var skipInfo map[string]interface{}
	childEvents = make([]map[string]map[string]string, 0)
	taskInfo.taskEidToName = make(map[int64]string)
	taskInfo.taskEidToTask = make(map[int64]string)
	taskInfo.taskEventID = make(map[string]int64)
	taskInfo.taskStartTime = make(map[string]int64)
	taskInfo.taskEndTime = make(map[string]int64)
	taskInfo.taskStatus = make(map[string]string)
	taskInfo.taskResult = make(map[string]interface{}) //taskResult的output字段是json格式
	taskInfo.taskErrorsMap = make(map[string]map[string]string)
	taskInfo.taskAttempt = make(map[string]string)
	taskInfo.taskIsCompleted = false
	taskInfo.taskIdMap = make([]map[int64]workflow.TaskID, 0)
	childSet := make(map[string]int)
	reciveSkipSign := false
	var skipTasks []string
	var firstExecutionRunId string
	for _, event := range historyEvents {
		//TODO:用ActivityTaskScheduled的historyEvents下标来区分是哪一个task，不用管decision，需要考虑任务的顺序都被打乱
		switch event.EventType.String() {
		case "DecisionTaskCompleted":
			reciveSkipSign = true //restart或者reset之后所有的sign都集中到最开始发送，以此条件过滤这些东西
		case "ActivityTaskScheduled":
			activityId := *event.ActivityTaskScheduledEventAttributes.ActivityId
			activityType := *event.ActivityTaskScheduledEventAttributes.ActivityType.Name
			taskInfo.taskEidToName[event.GetEventId()] = activityId
			taskInfo.taskEidToTask[event.GetEventId()] = activityType
			taskInfo.taskEventID[activityId] = *event.ActivityTaskScheduledEventAttributes.DecisionTaskCompletedEventId
			scheduleEvents = append(scheduleEvents, event)
			if _, exist := taskInfo.taskStatus[activityId]; exist {
				continue
			}
			taskInfo.taskStatus[activityId] = types.FlowNodeScheduleStatus
		case "ActivityTaskStarted":
			activityId := taskInfo.taskEidToName[*event.ActivityTaskStartedEventAttributes.ScheduledEventId]
			taskInfo.taskStartTime[activityId] = *event.Timestamp
			taskInfo.taskStatus[activityId] = types.FlowNodeStartStatus
			taskInfo.taskAttempt[activityId] = fmt.Sprint(*event.ActivityTaskStartedEventAttributes.Attempt)
		case "ActivityTaskCompleted":
			activityId := taskInfo.taskEidToName[*event.ActivityTaskCompletedEventAttributes.ScheduledEventId]
			taskInfo.taskEndTime[activityId] = *event.Timestamp
			taskInfo.taskStatus[activityId] = types.FlowNodeCompleteStatus
			if output := event.ActivityTaskCompletedEventAttributes.Result; output != nil {
				var outputMap interface{}
				err = json.Unmarshal(output, &outputMap)
				if err != nil {
					err = errors.New("Could not parse event.ActivityTaskCompletedEventAttributes:" + string(output))
					return
				} else {
					taskInfo.taskResult[activityId] = outputMap
				}
			}
		case "ActivityTaskTimedOut":
			activityId := taskInfo.taskEidToName[*event.ActivityTaskTimedOutEventAttributes.ScheduledEventId]
			taskInfo.taskStatus[activityId] = types.FlowNodeTimeOutStatus
			taskInfo.taskEndTime[activityId] = *event.Timestamp
			taskInfo.taskErrorsMap[activityId] = make(map[string]string)
			taskInfo.taskErrorsMap[activityId]["reason"] = "Timeout"
			taskInfo.taskErrorsMap[activityId]["details"] = event.ActivityTaskTimedOutEventAttributes.TimeoutType.String()
			owner, err := s.getOwner2(taskInfo.taskEidToTask[*event.ActivityTaskTimedOutEventAttributes.ScheduledEventId])
			if err != nil {
				err = errors.New("get task owner error:" + err.Error())
			}
			taskInfo.failedTaskOwner = owner
		case "ActivityTaskFailed":
			activityId := taskInfo.taskEidToName[*event.ActivityTaskFailedEventAttributes.ScheduledEventId]
			taskInfo.taskEndTime[activityId] = *event.Timestamp
			taskInfo.taskStatus[activityId] = types.FlowNodeFailStatus
			taskInfo.taskErrorsMap[activityId] = make(map[string]string)
			reason := string(event.ActivityTaskFailedEventAttributes.Details)
			if reason == "" {
				reason = *event.ActivityTaskFailedEventAttributes.Reason
			}
			taskInfo.taskErrorsMap[activityId]["details"] = reason
			taskInfo.taskErrorsMap[activityId]["reason"] = *event.ActivityTaskFailedEventAttributes.Reason
			owner, err := s.getOwner2(taskInfo.taskEidToTask[*event.ActivityTaskFailedEventAttributes.ScheduledEventId])
			if err != nil {
				err = errors.New("get task owner error:" + err.Error())
			}
			taskInfo.failedTaskOwner = owner
		case "WorkflowExecutionSignaled":
			if reciveSkipSign && *event.WorkflowExecutionSignaledEventAttributes.SignalName == workflow.SignalSkipTask {
				err = json.Unmarshal(event.WorkflowExecutionSignaledEventAttributes.Input, &skipInfo)
				if err != nil {
					return
				}
				for k, v := range skipInfo {
					//如果跳过的任务名不在整个流程中，则不处理，这里如果要跳过的任务是循环中的任务，会导致任务名会加上后缀，如果需要，后续再处理
					if _, ok := flowNodes[k]; !ok {
						continue
					}
					msg := v.(map[string]interface{})
					//使用skipTasks是因为跳过任务的信号可能很早就发送了，但scheduleEvents是按照任务执行的顺序加入的
					if _, exist := taskInfo.taskStatus[k]; !exist { //防止跳过已经scheduled的任务多次加入scheduleEvents
						skipTasks = append(skipTasks, k)
					}
					//跳过任务的attempt字段默认为0
					taskInfo.taskAttempt[k] = "0"
					if taskError := msg["error"]; taskError != "" && taskError != nil {
						taskInfo.taskStatus[k] = types.FlowNodeFailStatus
						taskInfo.taskErrorsMap[k] = make(map[string]string)
						taskInfo.taskErrorsMap[k]["reason"] = taskError.(string)
					} else {
						taskInfo.taskStatus[k] = types.FlowNodeSkipStatus
						taskInfo.taskResult[k] = msg["output"]
					}
					taskInfo.taskEndTime[k] = int64(time.Now().Nanosecond())
				}
			}
		//ActivityTask用taskRefName当作唯一标识，ChildEvent用workflowID当作唯一标识
		case "WorkflowExecutionCompleted":
			taskInfo.taskIsCompleted = true
		case "ChildWorkflowExecutionStarted":
			workflowId := *event.ChildWorkflowExecutionStartedEventAttributes.WorkflowExecution.WorkflowId
			runId := *event.ChildWorkflowExecutionStartedEventAttributes.WorkflowExecution.RunId
			childSet[workflowId] = len(childEvents)
			childEvents = append(childEvents, map[string]map[string]string{
				workflowId: {
					"runId":  runId,
					"status": types.FlowNodeStartStatus,
				},
			})
		case "ChildWorkflowExecutionCompleted":
			workflowId := *event.ChildWorkflowExecutionCompletedEventAttributes.WorkflowExecution.WorkflowId
			index := childSet[workflowId]
			childEvents[index][workflowId]["status"] = types.FlowNodeCompleteStatus
			childEvents[index][workflowId]["result"] = string((*event).ChildWorkflowExecutionCompletedEventAttributes.Result)
		case "ChildWorkflowExecutionFailed":
			workflowId := *event.ChildWorkflowExecutionFailedEventAttributes.WorkflowExecution.WorkflowId
			index := childSet[workflowId]
			childEvents[index][workflowId]["status"] = types.FlowNodeFailStatus
			childEvents[index][workflowId]["error"] = *event.ChildWorkflowExecutionFailedEventAttributes.Reason
		case "ChildWorkflowExecutionTimedOut":
			workflowId := *event.ChildWorkflowExecutionFailedEventAttributes.WorkflowExecution.WorkflowId
			index := childSet[workflowId]
			childEvents[index][workflowId]["status"] = types.FlowNodeTimeOutStatus
			childEvents[index][workflowId]["error"] = "timeout"
		case "DecisionTaskFailed":
			reciveSkipSign = false
			eventID := *event.EventId
			taskInfo.taskIdMap = append(taskInfo.taskIdMap, map[int64]workflow.TaskID{
				eventID: {
					WorkflowID: wfID,
					RunID:      *event.DecisionTaskFailedEventAttributes.BaseRunId,
				},
			})
		case "WorkflowExecutionStarted":
			firstExecutionRunId = *event.WorkflowExecutionStartedEventAttributes.FirstExecutionRunId
		}
	}
	if len(taskInfo.taskIdMap) > 0 {
		for k, _ := range taskInfo.taskIdMap[0] {
			taskInfo.taskIdMap[0][k] = workflow.TaskID{
				WorkflowID: wfID,
				RunID:      firstExecutionRunId,
			}
		}
	}
	return
}

func matchFormHistory(nodes *[]*types.FlowNode,
	taskInfo TaskInfo,
	scheduleEvents *[]*shared.HistoryEvent,
	getIndex *map[string]int,
	getFather *map[int]int,
	getDoWhile *map[string]int,
	wfID string,
	runID string) {
	pattern := regexp.MustCompile(`(.*)_(\d)$`)
	//todo:存在一个问题，如果任务的refName也是以 "_xxx" 结尾的，会出现错误。所以要坚决限制任务的refName

	for _, event := range *scheduleEvents {
		loopTime := 0
		activityOriginName := *(event.ActivityTaskScheduledEventAttributes.ActivityId)
		activityName := activityOriginName
		res := pattern.FindStringSubmatch(activityName)
		if res != nil {
			// 以下划线开头，未必就是loop，需要判断一下
			activityNamePrefix := res[1]
			_, exist := (*getIndex)[activityNamePrefix]
			if exist { //说明是loop
				activityName = activityNamePrefix
				loopTime, _ = strconv.Atoi(res[2])
			}
		}
		attr := event.ActivityTaskScheduledEventAttributes

		matchedNodeIndex, _ := (*getIndex)[activityName]
		matchedNode := (*nodes)[matchedNodeIndex]
		fatherNodeIndex := (*getFather)[matchedNodeIndex]
		fatherNode := (*nodes)[fatherNodeIndex]
		// 如果已经进入loop,则更新对应DO_WHILE的loopCount
		if loopTime != 0 {
			doWhileIndex := (*getDoWhile)[activityName]
			(*nodes)[doWhileIndex].LoopCount = loopTime
		}
		if fatherNode.DataType == decisionBranch { //如果父节点是decisionBranch，还要更新父节点的loopCount
			fatherNode.Status = "COMPLETED"
			fatherNode.LoopCount = loopTime
		}

		//stateChange
		// 开始写入
		// 更新，而不是直接append，所以要全都删除
		if loopTime != 0 {
			newConf := make([]*types.FlowNodeConf, 0)
			matchedNode.NodeConf = newConf
			activityName = activityOriginName //仍然是要用之前的ref-name来更新
		}
		taskID := workflow.TaskID{}
		taskID.WorkflowID = wfID
		taskID.RunID = runID
		taskID.Attempt = taskInfo.taskAttempt[matchedNode.RefName]
		taskID.TaskName = *attr.ActivityType.Name //node.Name
		taskID.ActivityID = *attr.ActivityId      //node.RefName
		matchedNode.TaskID = taskID.String()
		if startTime, ok := taskInfo.taskStartTime[matchedNode.RefName]; ok {
			tm := time.Unix(0, startTime)
			matchedNode.NodeConf = append(matchedNode.NodeConf, &types.FlowNodeConf{Label: "StartTime", Value: tm.In(timeLocation).String()})
		}
		if endTime, ok := taskInfo.taskEndTime[matchedNode.RefName]; ok {
			tm := time.Unix(0, endTime)
			matchedNode.NodeConf = append(matchedNode.NodeConf, &types.FlowNodeConf{Label: "EndTime", Value: tm.In(timeLocation).String()})
		}
		matchedNode.NodeConf = append(matchedNode.NodeConf, &types.FlowNodeConf{Label: "Input", Value: string(attr.Input)})
		matchedNode.NodeConf = append(matchedNode.NodeConf, &types.FlowNodeConf{Label: "Output", Value: taskInfo.taskResult[matchedNode.RefName]})
		matchedNode.NodeConf = append(matchedNode.NodeConf, &types.FlowNodeConf{Label: "ScheduleToCloseTimeoutSeconds", Value: attr.GetScheduleToCloseTimeoutSeconds()})
		matchedNode.NodeConf = append(matchedNode.NodeConf, &types.FlowNodeConf{Label: "ScheduleToStartTimeoutSeconds", Value: attr.GetScheduleToStartTimeoutSeconds()})
		matchedNode.NodeConf = append(matchedNode.NodeConf, &types.FlowNodeConf{Label: "StartToCloseTimeoutSeconds", Value: attr.GetStartToCloseTimeoutSeconds()})
		matchedNode.NodeConf = append(matchedNode.NodeConf, &types.FlowNodeConf{Label: "HeartbeatTimeoutSeconds", Value: attr.GetHeartbeatTimeoutSeconds()})
		matchedNode.NodeConf = append(matchedNode.NodeConf, &types.FlowNodeConf{Label: "DecisionTaskCompletedEventId", Value: attr.GetDecisionTaskCompletedEventId()})
		matchedNode.NodeConf = append(matchedNode.NodeConf, &types.FlowNodeConf{Label: "RetryPolicy", Value: attr.RetryPolicy})
		matchedNode.NodeConf = append(matchedNode.NodeConf, &types.FlowNodeConf{Label: "Header", Value: attr.Header})
		if taskError, ok := taskInfo.taskErrorsMap[matchedNode.RefName]; !ok {
			matchedNode.Status = taskInfo.taskStatus[matchedNode.RefName]
		} else {
			matchedNode.Status = taskInfo.taskStatus[matchedNode.RefName] //types.FlowNodeFailStatus
			matchedNode.NodeConf = append(matchedNode.NodeConf, &types.FlowNodeConf{Label: "Reasons", Value: taskError["details"]})
			matchedNode.NodeConf = append(matchedNode.NodeConf, &types.FlowNodeConf{Label: "ErrorType", Value: taskError["reason"]})
			matchedNode.NodeConf = append(matchedNode.NodeConf, &types.FlowNodeConf{Label: "Owner", Value: taskInfo.failedTaskOwner})
		}
	}
}

func getSimpleNodeIndex2(nodes *[]*types.FlowNode, edges *map[int]map[int]int, nextSimpleNodes *map[int]*map[string]int) {
	queue1 := list.New()
	queue2 := list.New()
	flowNodesLen := len(*nodes)
	var visited = make([]int, flowNodesLen, flowNodesLen)
	for i := 0; i < flowNodesLen; i++ {
		tmp := make(map[string]int)
		(*nextSimpleNodes)[i] = &tmp
	}
	queue1.PushBack(0)
	for queue1.Len() != 0 {
		nodeIndex := queue1.Remove(queue1.Front()).(int)
		visited[nodeIndex] = 1         //标记访问过
		nodeSet := (*edges)[nodeIndex] //当前nodeIndex的子节点
		for key, _ := range nodeSet {
			queue2.PushBack(key)
		}
		for queue2.Len() != 0 {
			subNodeIndex := queue2.Remove(queue2.Front()).(int)
			subNode := *(*nodes)[subNodeIndex] //当前子flowNode
			if subNode.DataType == "SIMPLE" || subNode.DataType == "HTTP" {
				if visited[subNodeIndex] != 1 {
					queue1.PushBack(subNodeIndex)
				}
				currSet := *(*nextSimpleNodes)[nodeIndex]
				currSet[subNode.RefName] = 0 //写表
			} else {
				nextNodesFromDull := (*edges)[subNodeIndex]
				for key, _ := range nextNodesFromDull {
					queue2.PushBack(key)
				}
			}
		}
	}
}
func mergeBtoA(setA *map[string]int, setB map[string]int) {
	for k, _ := range setB {
		(*setA)[k] = 0
	}
	return
}
func linkAllEdges(nodes *[]*types.FlowNode, edges *map[int]map[int]int, edgeOutputMap *map[string]*types.FlowEdge, visitCount *map[int]int, getDoWhile *map[string]int, lastTask string, getFater *map[int]int) {
	(*nodes)[0].Status = "COMPLETED" //将started节点 变为 完成
	queue := list.New()
	queue.PushBack(0)
	for queue.Len() != 0 {
		nodeIndex := queue.Remove(queue.Back()).(int)
		node := *(*nodes)[nodeIndex]
		subNodeIndex := (*edges)[nodeIndex]
		if node.DataType == "DECISION" { //因为decision每一次访问不是独立的，彼此之间会有影响，所以要拿出来单独考虑
			if node.Status == "COMPLETED" {
				isDefault := true //是否是default 的标志位
				loopCount := 0    //循环次数计数器
				var defaultNodeIndex int
				for k, _ := range subNodeIndex { //找到default节点index
					if (*nodes)[k].Name == "defaultCase" {
						defaultNodeIndex = k
						break
					}
				}
				//下面的for循环不更新default
				for k, _ := range subNodeIndex {
					subNode := (*nodes)[k]
					(*visitCount)[k]--
					if subNode.Status == "COMPLETED" {
						isDefault = false
						(*edgeOutputMap)[strconv.Itoa(nodeIndex)+"to"+strconv.Itoa(k)].Used = true
						newLoopCount := subNode.LoopCount
						loopCount += newLoopCount //不同decisionBranch节点的loopCount直接加起来就行
					}
					queue.PushBack(k) //为了避免发生一个循环走一个分支的情况，还是要不断遍历才行
				}
				//下面才开始更新default
				decisionLoopCount := (*nodes)[(*getDoWhile)[node.RefName]].LoopCount
				if isDefault || decisionLoopCount > loopCount { //如果确实是没有一个符合的，就是default
					(*nodes)[defaultNodeIndex].Status = "COMPLETED" //设置default节点为completed
					(*edgeOutputMap)[strconv.Itoa(nodeIndex)+"to"+strconv.Itoa(defaultNodeIndex)].Used = true
				}
				queue.PushBack(defaultNodeIndex)
			}
		} else {
			for k, _ := range subNodeIndex {
				canPush, _ := dealSubNode(nodes, nodeIndex, k, edgeOutputMap, visitCount, lastTask, getFater)
				if canPush {
					queue.PushBack(k)
				}
			}
		}
	}
}

func dealSubNode(nodes *[]*types.FlowNode, fatherNodeIndex int, subNodeIndex int, edgeOutputMap *map[string]*types.FlowEdge, visitCount *map[int]int, lastTask string, getFather *map[int]int) (canPush bool, isDefault bool) {
	grandFatherIndex := (*getFather)[fatherNodeIndex]
	grandFatherNode := (*nodes)[grandFatherIndex]
	fatherNode := (*nodes)[fatherNodeIndex]
	subNode := (*nodes)[subNodeIndex]
	(*visitCount)[subNodeIndex]--
	isDefault = true
	switch fatherNode.DataType {
	case "SIMPLE", "HTTP":
		if fatherNode.Status == "COMPLETED" || fatherNode.Status == "SKIPPED" {
			if subNode.DataType == "SIMPLE" || subNode.DataType == "HTTP" { // simple-->simple
				if subNode.Status == "COMPLETED" || subNode.Status == "SKIPPED" {
					(*edgeOutputMap)[strconv.Itoa(fatherNodeIndex)+"to"+strconv.Itoa(subNodeIndex)].Used = true
				}
			} else { //simple-->dull
				if subNode.DataType == "WAIT" {
					if fatherNode.RefName != lastTask {
						subNode.Status = "COMPLETED"
						(*edgeOutputMap)[strconv.Itoa(fatherNodeIndex)+"to"+strconv.Itoa(subNodeIndex)].Used = true
					}
				} else {
					subNode.Status = "COMPLETED"
					(*edgeOutputMap)[strconv.Itoa(fatherNodeIndex)+"to"+strconv.Itoa(subNodeIndex)].Used = true
				}
			}
		}
		if (*visitCount)[subNodeIndex] == 0 {
			canPush = true
		}
	case "DECISION":
	case "LOOP_END":
		if fatherNode.Status == "COMPLETED" || grandFatherNode.Status == "COMPLETED" {
			if subNodeIndex < fatherNodeIndex { // subNode是nodeIndex上面的节点，此时subNode必为DO_WHILE节点
				if subNode.LoopCount > 0 { //如果存在循环
					(*edgeOutputMap)[strconv.Itoa(fatherNodeIndex)+"to"+strconv.Itoa(subNodeIndex)].Used = true
					//不放入队列
				}
			} else { // subNode是nodeIndex下面的节点，此时subNode仍有可能是dull或者simple节点
				if subNode.DataType == "SIMPLE" || subNode.DataType == "HTTP" { // dull-->simple
					if subNode.Status == "COMPLETED" || subNode.Status == "SKIPPED" {
						(*edgeOutputMap)[strconv.Itoa(fatherNodeIndex)+"to"+strconv.Itoa(subNodeIndex)].Used = true
					}
				} else { // dull--->Dull
					subNode.Status = "COMPLETED" //如果上一个dull完成，那么他的子节点一定完成
					(*edgeOutputMap)[strconv.Itoa(fatherNodeIndex)+"to"+strconv.Itoa(subNodeIndex)].Used = true
				}
				if (*visitCount)[subNodeIndex] == 0 {
					canPush = true
				}
			}
		} else {
			//如果LOOP_END也失败，那他的子节点肯定全部失败
			if subNodeIndex > fatherNodeIndex { //不继续访问上方DO_WHILE节点
				if (*visitCount)[subNodeIndex] == 0 {
					canPush = true
				}
			}
		}
	default:
		if fatherNode.Status == "COMPLETED" {
			if subNode.DataType == "SIMPLE" || subNode.DataType == "HTTP" { // dull-->simple
				if subNode.Status == "COMPLETED" || subNode.Status == "SKIPPED" {
					(*edgeOutputMap)[strconv.Itoa(fatherNodeIndex)+"to"+strconv.Itoa(subNodeIndex)].Used = true
				}
			} else { // dull--->dull
				subNode.Status = "COMPLETED" //如果上一个dull完成，那么他的子节点一定完成
				(*edgeOutputMap)[strconv.Itoa(fatherNodeIndex)+"to"+strconv.Itoa(subNodeIndex)].Used = true
			}
		}
		if (*visitCount)[subNodeIndex] == 0 {
			canPush = true
		}
	}
	return
}






*/
