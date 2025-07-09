package repository

import "database/sql"

type People struct {
	ID    int
	Name  string
	Day   int
	Month int
}

type PeopleRepo interface {
	Get(day, month int) ([]People, error)
}

type PeopleRepoImpl struct {
	DB *sql.DB
}

func (r *PeopleRepoImpl) Get(day, month int) ([]People, error) {
	rows, err := r.DB.Query("SELECT id, name, day, month FROM people WHERE day = $1 AND month = $2", day, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []People
	for rows.Next() {
		var p People
		if err := rows.Scan(&p.ID, &p.Name, &p.Day, &p.Month); err != nil {
			return nil, err
		}
		people = append(people, p)
	}
	return people, nil
}
