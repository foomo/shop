package order

// func (order *Order) SaveOrderEvent(action ActionOrder, err error, description string) {
// 	debug.Log("Action", string(action), "OrderID", order.Id)
// 	event := event_log.NewEvent()
// 	if err != nil {
// 		event.Type = event_log.EventTypeError
// 	} else {
// 		event.Type = event_log.EventTypeSuccess
// 	}
// 	event.Action = string(action)
// 	event.OrderID = order.Id
// 	event.Description = description
// 	if err != nil {
// 		event.Error = err.Error()
// 	}
// 	order.History = append(order.History, event)
// 	order.Upsert() // Error is ignored because it gets already logged in UpsertOrder()

// 	jsonBytes, _ := json.MarshalIndent(event, "", "	")
// 	debug.Log("Saved Order Event! ", string(jsonBytes))
// }

// // Event will only be saved if is an error
// func (order *Order) SaveOrderEventOnError(action ActionOrder, err error, description string) {
// 	if err == nil {
// 		return
// 	}
// 	order.SaveOrderEvent(action, err, description)
// }

// func (order *Order) SaveOrderEventCustomEvent(e event_log.Event) {
// 	order.History = append(order.History, &e)
// 	order.Upsert() // Error is ignored because it gets already logged in UpsertOrder()
// 	jsonBytes, _ := json.MarshalIndent(&e, "", "	")
// 	debug.Log("Saved Order Event! ", string(jsonBytes))
// }

// func (order *Order) ReportErrors(printOnConsole bool) string {
// 	errCount := 0
// 	if len(order.History) > 0 {
// 		errCount++
// 		jsonBytes, err := json.MarshalIndent(order.History, "", "	")
// 		if err != nil {
// 			panic(err)
// 		}
// 		s := string(jsonBytes)
// 		if printOnConsole {
// 			log.Println("Errors logged for order with orderID:")
// 			log.Println(s)
// 		}

// 		return s
// 	}
// 	return "No errors logged for order with orderID " + order.Id
// }
