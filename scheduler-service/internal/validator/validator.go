package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/namnv2496/scheduler/internal/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IValidate interface {
	ValidateRequire(ctx context.Context, action string, eventMap map[string]string) error
	ValidateValue(ctx context.Context, eventMap map[string]string) error
	ValidateCustomeRules(eventMap map[string]string) error
}

type Validate struct {
	validateByActions map[string]entity.FieldValidation
	requireByActions  map[string]map[string]entity.FieldRequireCondition
	customValidators  map[string][]entity.CrossFieldRule
}

func NewValidate() *Validate {
	return &Validate{
		validateByActions: make(map[string]entity.FieldValidation),
		requireByActions:  make(map[string]map[string]entity.FieldRequireCondition),
		customValidators:  registerCustomeRules(),
	}
}

func (_self *Validate) ValidateRequire(ctx context.Context, action string, eventMap map[string]string) error {
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
	if err := validateRequireEventInfo(eventMap, requireByAction); err != nil {
		return err
	}
	return nil
}

func (_self *Validate) ValidateValue(ctx context.Context, eventMap map[string]string) error {
	if len(_self.validateByActions) == 0 {
		requires, err := getRules()
		if err != nil {
			return err
		}
		_self.validateByActions = requires
	}
	if err := validateValueEventInfo(eventMap, _self.validateByActions); err != nil {
		return err
	}
	return nil
}

func (_self *Validate) ValidateCustomeRules(eventMap map[string]string) error {
	for paramName, value := range eventMap {
		if err := _self.validateCustomeRules(paramName, value, eventMap); err != nil {
			return err
		}
	}
	return nil
}

func (_self *Validate) validateCustomeRules(paramName, value string, eventFields map[string]string) error {
	rules, exist := _self.customValidators[paramName]
	if !exist {
		return nil
	}

	for _, rule := range rules {
		if err := validateFieldCustomeRule(paramName, value, rule, eventFields); err != nil {
			return err
		}
	}
	return nil
}

func validateFieldCustomeRule(paramName, value string, rule entity.CrossFieldRule, eventFields map[string]string) error {
	var valid bool
	var err error

	if rule.Pattern != "" {
		if value != "" {
			matched, err := regexp.MatchString(rule.Pattern, value)
			if err != nil {
				return fmt.Errorf("%s: Lỗi pattern không hợp lệ", paramName)
			}
			if !matched {
				if rule.ErrorMsg != "" {
					return fmt.Errorf("%s: %s", paramName, rule.ErrorMsg)
				}
				return fmt.Errorf("%s: Giá trị không đúng định dạng", paramName)
			}
		}
	}

	// AllowedValues (Enum) validation
	if len(rule.AllowedValues) > 0 {
		if value != "" {
			found := slices.Contains(rule.AllowedValues, value)
			if !found {
				return fmt.Errorf("%s: Giá trị không hợp lệ. Chỉ chấp nhận: %s", paramName, strings.Join(rule.AllowedValues, ", "))
			}
		}
	}
	if len(rule.Operator) > 0 {
		compareValue := eventFields[rule.Field]
		if len(rule.Value) > 0 {
			compareValue = rule.Value
		}

		switch rule.Operator {
		case entity.OP_EQ: // equal
			valid = value == compareValue
		case entity.OP_NE: // not equal
			valid = value != compareValue
		case entity.OP_GT: // greater than
			valid, err = compareNumeric(value, compareValue, func(a, b float64) bool { return a > b })
		case entity.OP_GTE: // greater than or equal
			valid, err = compareNumeric(value, compareValue, func(a, b float64) bool { return a >= b })
		case entity.OP_LT: // less than
			valid, err = compareNumeric(value, compareValue, func(a, b float64) bool { return a < b })
		case entity.OP_LTE: // less than or equal
			valid, err = compareNumeric(value, compareValue, func(a, b float64) bool { return a <= b })
		default:
			return fmt.Errorf("%s: Toán tử không hợp lệ '%s'", paramName, rule.Operator)
		}

		if err != nil {
			return fmt.Errorf("%s: %v", paramName, err)
		}
		if !valid {
			if rule.ErrorMsg != "" {
				return fmt.Errorf("%s: %s", paramName, rule.ErrorMsg)
			}
		}
	}

	return nil
}

// compareNumeric compares two string values as numbers using the provided comparison function
func compareNumeric(a, b string, compare func(float64, float64) bool) (bool, error) {
	if a == "" || b == "" {
		return false, fmt.Errorf("giá trị rỗng không thể so sánh")
	}

	aNum, err := strconv.ParseFloat(a, 64)
	if err != nil {
		return false, fmt.Errorf("giá trị '%s' không phải là số", a)
	}

	bNum, err := strconv.ParseFloat(b, 64)
	if err != nil {
		return false, fmt.Errorf("giá trị '%s' không phải là số", b)
	}

	return compare(aNum, bNum), nil
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

func validateRequireEventInfo(eventMap map[string]string, rules map[string]entity.FieldRequireCondition) error {
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

func validateValueEventInfo(eventMap map[string]string, valueCondition map[string]entity.FieldValidation) error {
	for paramName, condition := range valueCondition {
		if err := validateFieldValue(eventMap, paramName, condition); err != nil {
			return err
		}
	}
	return nil
}

func validateFieldValue(eventMap map[string]string, paramName string, valueCondition entity.FieldValidation) error {
	inputValue := eventMap[paramName]
	if valueCondition.MinValue > 0 {
		if inputValue != "" {
			value, err := strconv.Atoi(inputValue)
			if err != nil {
				return err
			}
			if value < valueCondition.MinValue {
				return fmt.Errorf("\"%s\" is required >= %d", valueCondition.Label, valueCondition.MinValue)
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
				return fmt.Errorf("\"%s\" is required <= %d", valueCondition.Label, valueCondition.MaxValue)
			}
		}
	}
	if valueCondition.MinLength > 0 {
		if inputValue != "" {
			if len(inputValue) < valueCondition.MinLength {
				return fmt.Errorf("\"%s\" is required >= %d characters", valueCondition.Label, valueCondition.MinLength)
			}
		}
	}
	if valueCondition.MaxLength > 0 {
		if inputValue != "" {
			if len(inputValue) > valueCondition.MaxLength {
				return fmt.Errorf("\"%s\" is required <= %d characters", valueCondition.Label, valueCondition.MaxLength)
			}
		}
	}
	if valueCondition.MinWord > 0 {
		if inputValue != "" {
			words := strings.Fields(inputValue)
			if int(len(words)) < valueCondition.MinWord {
				return fmt.Errorf("\"%s\" is required >= %d words", valueCondition.Label, valueCondition.MinWord)
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

func registerCustomeRules() map[string][]entity.CrossFieldRule {
	customValidators := make(map[string][]entity.CrossFieldRule)
	rules := map[string][]entity.CrossFieldRule{
		"method": {
			{
				AllowedValues: []string{"GET", "POST"},
			},
		},
		"repeat_times": {
			{
				Operator: entity.OP_LTE,
				Value:    "1",
				ErrorMsg: "Số lần lặp tối thiểu >= 1",
			},
			{
				Operator: entity.OP_GT,
				Value:    "1000",
				ErrorMsg: "Số lần lặp tối thiểu  < 1000",
			},
		},
		"description": {
			{
				Value:    "2",
				Operator: entity.OP_GTE,
				ErrorMsg: "Độ dài text phải >= 2",
			},
		},
		"scheduler_at": {
			{
				Field:    "next_run_time",
				Operator: entity.OP_EQ,
				ErrorMsg: "Thời gian trigger lần đầu phải trùng với next_run_time",
			},
			{
				Pattern:  `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`,
				ErrorMsg: "Thời gian trigger không đúng format",
			},
		},
	}
	maps.Copy(customValidators, rules)
	return customValidators
}
