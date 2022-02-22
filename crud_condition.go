package mog

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

/**
 * @project momo
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-30 11:49
 * @desc
 */
type Condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

type Conditions []Condition

func EmptyConditions() Conditions {
	return Conditions{}
}

func (c Conditions) IsZero() bool {
	return len(Conditions{}) > 0
}

func NewEqualsCondition(field string, value any) Condition {
	return NewCondition(field, OPERATOR_EQUAL, value)
}

func NewCondition(field, operator string, value any) Condition {
	return Condition{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

func SingleEqualsConditions(field string, value any) Conditions {
	return SingleConditions(field, OPERATOR_EQUAL, value)
}

func SingleConditions(field, operator string, value any) Conditions {
	return Conditions{
		Condition{
			Field:    field,
			Operator: operator,
			Value:    value,
		},
	}
}

const OPERATOR_LIKE = "like"
const OPERATOR_NOT_LIKE = "notlike"
const OPERATOR_EQUAL = "eq"
const OPERATOR_NOT_EQUAL = "ne"
const OPERATOR_GREATER_THAN = "gt"
const OPERATOR_GREATER_THAN_OR_EAUAL = "ge"
const OPERATOR_LESS_THAN = "lt"
const OPERATOR_LESS_THAN_OR_EQUAL = "le"
const OPERATOR_IN = "in"
const OPERATOR_NOT_IN = "notin"
const OPERATOR_BETWEEN = "between"

func WithConditions(db **gorm.DB, conditions Conditions) error {
	if !conditions.IsZero() {
		for _, v := range conditions {
			switch strings.ToLower(v.Operator) {
			case OPERATOR_LIKE:
				*db = (*db).Where(fmt.Sprintf("%s LIKE ?", v.Field), "%"+fmt.Sprintf("%v", v.Value)+"%")
			case OPERATOR_NOT_LIKE:
				*db = (*db).Where(fmt.Sprintf("%s NOT LIKE ?", v.Field), "%"+fmt.Sprintf("%v", v.Value)+"%")
			case OPERATOR_EQUAL: // =
				*db = (*db).Where(fmt.Sprintf("%s=?", v.Field), v.Value)
			case OPERATOR_NOT_EQUAL: // !=
				*db = (*db).Where(fmt.Sprintf("%s!=?", v.Field), v.Value)
			case OPERATOR_GREATER_THAN: // >
				*db = (*db).Where(fmt.Sprintf("%s>?", v.Field), v.Value)
			case OPERATOR_GREATER_THAN_OR_EAUAL: // >=
				*db = (*db).Where(fmt.Sprintf("%s>=?", v.Field), v.Value)
			case OPERATOR_LESS_THAN: // <
				*db = (*db).Where(fmt.Sprintf("%s<?", v.Field), v.Value)
			case OPERATOR_LESS_THAN_OR_EQUAL: // <=
				*db = (*db).Where(fmt.Sprintf("%s<=?", v.Field), v.Value)
			case OPERATOR_IN: // in
				val, ok := v.Value.([]any)
				if !ok {
					return fmt.Errorf("condition %s must be a list", v.Field)
				}
				*db = (*db).Where(fmt.Sprintf("%s IN (?)", v.Field), val)
			case OPERATOR_NOT_IN: // notin
				val, ok := v.Value.([]any)
				if !ok {
					return fmt.Errorf("condition %s must be a list", v.Field)
				}
				*db = (*db).Where(fmt.Sprintf("%s NOT IN (?)", v.Field), val)
			case OPERATOR_BETWEEN: //
				val, ok := v.Value.([]any)
				if !ok {
					return fmt.Errorf("condition %s must be a list", v.Field)
				}
				if !(len(val) == 2) {
					return fmt.Errorf("condition %s length must be 2", v.Field)
				}
				*db = (*db).Where(fmt.Sprintf("%s BETWEEN ? AND ?", v.Field), val[0], val[1])
			}
		}
	}
	return nil
}

func WithOrConditions(db **gorm.DB, conditions Conditions) error {
	if !conditions.IsZero() {
		for _, v := range conditions {
			switch strings.ToLower(v.Operator) {
			case OPERATOR_LIKE:
				*db = (*db).Or(fmt.Sprintf("%s LIKE ?", v.Field), "%"+fmt.Sprintf("%v", v.Value)+"%")
			case OPERATOR_NOT_LIKE:
				*db = (*db).Or(fmt.Sprintf("%s NOT LIKE ?", v.Field), "%"+fmt.Sprintf("%v", v.Value)+"%")
			case OPERATOR_EQUAL: // =
				*db = (*db).Or(fmt.Sprintf("%s=?", v.Field), v.Value)
			case OPERATOR_NOT_EQUAL: // !=
				*db = (*db).Or(fmt.Sprintf("%s!=?", v.Field), v.Value)
			case OPERATOR_GREATER_THAN: // >
				*db = (*db).Or(fmt.Sprintf("%s>?", v.Field), v.Value)
			case OPERATOR_GREATER_THAN_OR_EAUAL: // >=
				*db = (*db).Or(fmt.Sprintf("%s>=?", v.Field), v.Value)
			case OPERATOR_LESS_THAN: // <
				*db = (*db).Or(fmt.Sprintf("%s<?", v.Field), v.Value)
			case OPERATOR_LESS_THAN_OR_EQUAL: // <=
				*db = (*db).Or(fmt.Sprintf("%s<=?", v.Field), v.Value)
			case OPERATOR_IN: // in
				val, ok := v.Value.([]any)
				if !ok {
					return fmt.Errorf("condition %s must be a list", v.Field)
				}
				*db = (*db).Or(fmt.Sprintf("%s IN (?)", v.Field), val)
			case OPERATOR_NOT_IN: // notin
				val, ok := v.Value.([]any)
				if !ok {
					return fmt.Errorf("condition %s must be a list", v.Field)
				}
				*db = (*db).Or(fmt.Sprintf("%s NOT IN (?)", v.Field), val)
			case OPERATOR_BETWEEN: //
				val, ok := v.Value.([]any)
				if !ok {
					return fmt.Errorf("condition %s must be a list", v.Field)
				}
				if !(len(val) == 2) {
					return fmt.Errorf("condition %s length must be 2", v.Field)
				}
				*db = (*db).Or(fmt.Sprintf("%s BETWEEN ? AND ?", v.Field), val[0], val[1])
			}
		}
		*db = (*db).Where(*db)
	}
	return nil
}
