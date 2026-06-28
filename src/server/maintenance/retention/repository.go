package retention

import (
	"fmt"
	"server/configs"
	"server/tools"
	"strconv"
)

type RetentionPolicyRepository interface {
	GetRetentionPolicy() (*RetentionPolicy, error)
	IsRetentionPolicySet() (bool, error)
	SetRetentionPolicy(*RetentionPolicy) error
}

type RetentionPolicyRepositoryImpl struct {
	DatabaseConnector tools.DatabaseConnector
}

func (r *RetentionPolicyRepositoryImpl) GetRetentionPolicy() (*RetentionPolicy, error) {
	var preUpdateRetentionString string
	var dailyRetentionString string
	var weeklyRetentionString string
	var monthlyRetentionString string
	var yearlyRetentionString string

	err := r.DatabaseConnector.GetDB().QueryRow(`
SELECT
	(SELECT value FROM configs WHERE key = $1),
	(SELECT value FROM configs WHERE key = $2),
	(SELECT value FROM configs WHERE key = $3),
	(SELECT value FROM configs WHERE key = $4),
	(SELECT value FROM configs WHERE key = $5);
`,
		configs.ConfigKeys.BackupRetentionPreUpdated,
		configs.ConfigKeys.BackupRetentionDaily,
		configs.ConfigKeys.BackupRetentionWeekly,
		configs.ConfigKeys.BackupRetentionMonthly,
		configs.ConfigKeys.BackupRetentionYearly,
	).Scan(
		&preUpdateRetentionString,
		&dailyRetentionString,
		&weeklyRetentionString,
		&monthlyRetentionString,
		&yearlyRetentionString,
	)
	if err != nil {
		return nil, err
	}

	preUpdateRetentionCount, err := strconv.Atoi(preUpdateRetentionString)
	if err != nil {
		return nil, err
	}
	dailyRetentionCount, err := strconv.Atoi(dailyRetentionString)
	if err != nil {
		return nil, err
	}
	weeklyRetentionCount, err := strconv.Atoi(weeklyRetentionString)
	if err != nil {
		return nil, err
	}
	monthlyRetentionCount, err := strconv.Atoi(monthlyRetentionString)
	if err != nil {
		return nil, err
	}
	yearlyRetentionCount, err := strconv.Atoi(yearlyRetentionString)
	if err != nil {
		return nil, err
	}

	return &RetentionPolicy{
		KeepPreUpdate: preUpdateRetentionCount,
		KeepDaily:     dailyRetentionCount,
		KeepWeekly:    weeklyRetentionCount,
		KeepMonthly:   monthlyRetentionCount,
		KeepYearly:    yearlyRetentionCount,
	}, nil
}

func (r *RetentionPolicyRepositoryImpl) IsRetentionPolicySet() (bool, error) {
	var count int
	err := r.DatabaseConnector.GetDB().QueryRow(`
SELECT COUNT(*) FROM configs
WHERE key IN ($1, $2, $3, $4, $5);
`,
		configs.ConfigKeys.BackupRetentionPreUpdated,
		configs.ConfigKeys.BackupRetentionDaily,
		configs.ConfigKeys.BackupRetentionWeekly,
		configs.ConfigKeys.BackupRetentionMonthly,
		configs.ConfigKeys.BackupRetentionYearly,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 5, nil
}

func (r *RetentionPolicyRepositoryImpl) SetRetentionPolicy(policy *RetentionPolicy) error {
	_, err := r.DatabaseConnector.GetDB().Exec(`
INSERT INTO configs (key, value)
VALUES
	($1, $2),
	($3, $4),
	($5, $6),
	($7, $8),
	($9, $10)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value;
`,
		configs.ConfigKeys.BackupRetentionPreUpdated, fmt.Sprintf("%d", policy.KeepPreUpdate),
		configs.ConfigKeys.BackupRetentionDaily, fmt.Sprintf("%d", policy.KeepDaily),
		configs.ConfigKeys.BackupRetentionWeekly, fmt.Sprintf("%d", policy.KeepWeekly),
		configs.ConfigKeys.BackupRetentionMonthly, fmt.Sprintf("%d", policy.KeepMonthly),
		configs.ConfigKeys.BackupRetentionYearly, fmt.Sprintf("%d", policy.KeepYearly),
	)
	return err
}

func GetDefaultRetentionPolicy() *RetentionPolicy {
	return &RetentionPolicy{
		KeepPreUpdate: 5,
		KeepDaily:     7,
		KeepWeekly:    4,
		KeepMonthly:   12,
		KeepYearly:    2,
	}
}
