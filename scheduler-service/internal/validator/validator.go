package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/namnv2496/scheduler/internal/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IValidate interface {
	ValidateRequire(ctx context.Context, action string, req *entity.CrawlerEvent) error
	ValidateValue(ctx context.Context, req *entity.CrawlerEvent) error
}

type Validate struct {
	validateByActions map[string]entity.FieldValidation
	requireByActions  map[string]map[string]entity.FieldRequireCondition
}

func NewValidate() *Validate {
	return &Validate{
		validateByActions: make(map[string]entity.FieldValidation),
		requireByActions:  make(map[string]map[string]entity.FieldRequireCondition),
	}
}

func (_self *Validate) ValidateRequire(ctx context.Context, action string, req *entity.CrawlerEvent) error {
	var requireByAction map[string]entity.FieldRequireCondition
	var exist bool
	requireByAction, exist = _self.requireByActions[action]
	if !exist {
		requires, err := getRequireConditions()
		if err != nil {
			return err
		}
		_self.requireByActions = requires
		requireByAction = _self.requireByActions[action]
	}
	if err := validateRequireEventInfo(req, requireByAction); err != nil {
		return err
	}
	return nil
}

func (_self *Validate) ValidateValue(ctx context.Context, req *entity.CrawlerEvent) error {
	if len(_self.validateByActions) == 0 {
		requires, err := getRules()
		if err != nil {
			return err
		}
		_self.validateByActions = requires
	}
	if err := validateValueEventInfo(req, _self.validateByActions); err != nil {
		return err
	}
	return nil
}

func getRequireConditions() (map[string]map[string]entity.FieldRequireCondition, error) {
	data, err := os.ReadFile("./internal/validator/requires.yaml")
	if err != nil {
		return nil, err
	}
	var requirements map[string]map[string]entity.FieldRequire
	err = json.Unmarshal(data, &requirements)
	if err != nil {
		return nil, err
	}
	resp := make(map[string]map[string]entity.FieldRequireCondition)

	for action, rules := range requirements {
		updateRules := make(map[string]entity.FieldRequireCondition, 0)
		for paramName, conditions := range rules {
			conditionRules := make([]string, 0)
			for dependParam, dependValue := range conditions {
				conditionRules = append(conditionRules, dependParam+":"+dependValue)
			}
			updateRules[paramName] = entity.FieldRequireCondition{
				Param:      paramName,
				Name:       paramName, // update displayedName later
				Conditions: conditionRules,
			}
		}
		resp[action] = updateRules
	}
	return resp, nil
}

func getRules() (map[string]entity.FieldValidation, error) {
	data, err := os.ReadFile("./internal/validator/rules.yaml")
	if err != nil {
		return nil, err
	}
	var requirements map[string]entity.FieldValidation
	err = json.Unmarshal(data, &requirements)
	if err != nil {
		return nil, err
	}
	return requirements, nil
}

func validateRequireEventInfo(event *entity.CrawlerEvent, rules map[string]entity.FieldRequireCondition) error {
	eventMap := event.ToMap()
	for paramName, requireCondition := range rules {
		if err := validateRequireField(eventMap, paramName, requireCondition); err != nil {
			return err
		}
	}
	return nil
}

func validateRequireField(eventMap map[string]string, paramName string, requireCondition entity.FieldRequireCondition) error {
	inputValue := eventMap[paramName]
	for _, condition := range requireCondition.Conditions {
		parts := strings.Split(condition, ":")
		// check require = 1
		if parts[0] == "require" && parts[1] == "1" && inputValue == "" {
			return status.Errorf(codes.InvalidArgument, "Vui lòng nhập thông tin \"%s\"", requireCondition.Name)
		} else if parts[0] != "require" {
			// check depend field first
			if eventMap[parts[0]] == parts[1] {
				// check current field
				if inputValue == "" {
					return status.Errorf(codes.InvalidArgument, "Vui lòng nhập thông tin \"%s\"", requireCondition.Name)
				}
			}
		}
	}
	return nil
}

func validateValueEventInfo(event *entity.CrawlerEvent, valueCondition map[string]entity.FieldValidation) error {
	eventMap := event.ToMap()
	for paramName, condition := range valueCondition {
		if err := validateValueField(eventMap, paramName, condition); err != nil {
			return err
		}
	}
	return nil
}

func validateValueField(eventMap map[string]string, paramName string, valueCondition entity.FieldValidation) error {
	inputValue := eventMap[paramName]
	if valueCondition.MinValue > 0 {
		if inputValue != "" {
			value, err := strconv.Atoi(inputValue)
			if err != nil {
				return err
			}
			if value < valueCondition.MinValue {
				return fmt.Errorf("%s is required >= %d", paramName, valueCondition.MinValue)
			}
		}
	}
	if valueCondition.MaxValue > 0 {
		if inputValue != "" {
			value, err := strconv.Atoi(inputValue)
			if err != nil {
				return err
			}
			if value > valueCondition.MaxValue {
				return fmt.Errorf("%s is required <= %d", paramName, valueCondition.MaxValue)
			}
		}
	}
	if valueCondition.MinLength > 0 {
		if inputValue != "" {
			if len(inputValue) < valueCondition.MinLength {
				return fmt.Errorf("%s is required >= %d characters", paramName, valueCondition.MinLength)
			}
		}
	}
	if valueCondition.MaxLength > 0 {
		if inputValue != "" {
			if len(inputValue) > valueCondition.MaxLength {
				return fmt.Errorf("%s is required <= %d characters", paramName, valueCondition.MaxLength)
			}
		}
	}
	if valueCondition.MinWord > 0 {
		if inputValue != "" {
			words := strings.Fields(inputValue)
			if int(len(words)) < valueCondition.MinWord {
				return fmt.Errorf("%s is required >= %d words", paramName, valueCondition.MinWord)
			}

		}
	}
	if valueCondition.MaxWord > 0 {
		if inputValue != "" {
			words := strings.Fields(inputValue)
			if int(len(words)) > valueCondition.MaxWord {
				return fmt.Errorf("%s is required <= %d words", paramName, valueCondition.MaxWord)
			}

		}
	}
	return nil
}
