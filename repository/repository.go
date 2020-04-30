package repository

import "github.com/jmoiron/sqlx"

type a2billingRepository struct {
	db *sqlx.DB
}

func (r a2billingRepository) GetAgentIdsByParentAgentID(id int32) (result []int32, err error) {
	rows, err := r.db.Queryx("SELECT id FROM cc_agent WHERE parent_agent_id = ?", id)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var agentID int32
		err = rows.Scan(&agentID)
		if err != nil {
			return
		}
		result = append(result, agentID)
	}
	return
}

func NewA2billingRepository(db *sqlx.DB) A2billingRepository {
	return &a2billingRepository{db}
}
