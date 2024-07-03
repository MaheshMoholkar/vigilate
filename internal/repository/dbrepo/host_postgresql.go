package dbrepo

import (
	"context"
	"log"
	"time"

	"github.com/tsawler/vigilate/internal/models"
)

// InsertHost inserts a host into the database
func (m *postgresDBRepo) InsertHost(h models.Host) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into hosts (host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning id`

	var newID int

	err := m.DB.QueryRowContext(ctx, query,
		h.HostName,
		h.CanonicalName,
		h.URL,
		h.IP,
		h.IPV6,
		h.Location,
		h.OS,
		h.Active,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		log.Println(err)
		return 0, err
	}

	// add host services and set to inactive
	query = `select id from services`
	serviceRows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer serviceRows.Close()

	for serviceRows.Next() {
		var serviceID int
		err = serviceRows.Scan(&serviceID)
		if err != nil {
			log.Println(err)
			return 0, err
		}

		stmt := `insert into host_services (host_id, service_id, active, schedule_number, schedule_unit, status, created_at, updated_at) values ($1, $2, 0, 3, 'm', 'pending', $3, $4)`

		_, err = m.DB.ExecContext(ctx, stmt, newID, serviceID, time.Now(), time.Now())
		if err != nil {
			log.Println(err)
			return newID, err
		}
	}

	return newID, nil
}

// GetHostByID gets a host by id
func (m *postgresDBRepo) GetHostByID(id int) (models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at from hosts where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var h models.Host

	err := row.Scan(
		&h.ID,
		&h.HostName,
		&h.CanonicalName,
		&h.URL,
		&h.IP,
		&h.IPV6,
		&h.Location,
		&h.OS,
		&h.Active,
		&h.CreatedAt,
		&h.UpdatedAt,
	)

	if err != nil {
		return h, err
	}

	// get all service for host
	query = `select hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit, hs.last_check, hs.status, hs.created_at, hs.updated_at, s.id, s.service_name, s.icon, s.created_at, s.updated_at from host_services hs left join services s on (hs.service_id = s.id) where hs.host_id = $1 order by s.service_name`

	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return h, err
	}
	defer rows.Close()

	var hostServices []models.HostService

	for rows.Next() {
		var hs models.HostService
		err = rows.Scan(
			&hs.ID,
			&hs.HostID,
			&hs.ServiceID,
			&hs.Active,
			&hs.ScheduleNumber,
			&hs.ScheduleUnit,
			&hs.LastCheck,
			&hs.Status,
			&hs.CreatedAt,
			&hs.UpdatedAt,
			&hs.Service.ID,
			&hs.Service.ServiceName,
			&hs.Service.Icon,
			&hs.Service.CreatedAt,
			&hs.Service.UpdatedAt,
		)
		if err != nil {
			return h, err
		}
		hostServices = append(hostServices, hs)
	}

	if err = rows.Err(); err != nil {
		return h, err
	}

	h.HostServices = hostServices
	return h, nil
}

// UpdateHost updates a host in the database
func (m *postgresDBRepo) UpdateHost(h models.Host) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update hosts set host_name = $1, canonical_name = $2, url = $3, ip = $4, ipv6 = $5, location = $6, os = $7, active = $8, updated_at = $9 where id = $10`

	_, err := m.DB.ExecContext(ctx, stmt,
		h.HostName,
		h.CanonicalName,
		h.URL,
		h.IP,
		h.IPV6,
		h.Location,
		h.OS,
		h.Active,
		time.Now(),
		h.ID,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// AllHost returns all hosts
func (m *postgresDBRepo) AllHosts() ([]*models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at from hosts order by host_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []*models.Host

	for rows.Next() {
		var h models.Host
		err = rows.Scan(
			&h.ID,
			&h.HostName,
			&h.CanonicalName,
			&h.URL,
			&h.IP,
			&h.IPV6,
			&h.Location,
			&h.OS,
			&h.Active,
			&h.CreatedAt,
			&h.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		serviceQuery := `
			select 
				hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit, hs.last_check, hs.status, hs.created_at, hs.updated_at, s.id, s.service_name, s.icon, s.created_at, s.updated_at, hs.last_message 
			from 
				host_services hs 
				left join services s on (hs.service_id = s.id) 
			where 
				hs.host_id = $1 order by s.service_name`

		serviceRows, err := m.DB.QueryContext(ctx, serviceQuery, h.ID)
		if err != nil {
			return nil, err
		}

		var hostServices []models.HostService

		for serviceRows.Next() {
			var hs models.HostService
			err = serviceRows.Scan(
				&hs.ID,
				&hs.HostID,
				&hs.ServiceID,
				&hs.Active,
				&hs.ScheduleNumber,
				&hs.ScheduleUnit,
				&hs.LastCheck,
				&hs.Status,
				&hs.CreatedAt,
				&hs.UpdatedAt,
				&hs.Service.ID,
				&hs.Service.ServiceName,
				&hs.Service.Icon,
				&hs.Service.CreatedAt,
				&hs.Service.UpdatedAt,
				&hs.LastMessage,
			)
			if err != nil {
				return nil, err
			}
			hostServices = append(hostServices, hs)
			serviceRows.Close()
		}

		if err = serviceRows.Err(); err != nil {
			return nil, err
		}

		h.HostServices = hostServices

		hosts = append(hosts, &h)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return hosts, nil
}

// UpdateHostServiceStatus updates the status of a host service
func (m *postgresDBRepo) UpdateHostServiceStatus(hostID, serviceID, active int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update host_services set active = $1 where host_id = $2 and service_id = $3`

	_, err := m.DB.ExecContext(ctx, stmt, active, hostID, serviceID)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// UpdateHostServiceStatus updates the status of a host service
func (m *postgresDBRepo) UpdateHostService(hs models.HostService) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update 
				host_services set
				    host_id = $1, service_id = $2, active = $3, schedule_number = $4, schedule_unit = $5, last_check = $6, status = $7, updated_at = $8, last_message = $9
				where
					id = $10`

	_, err := m.DB.ExecContext(ctx, stmt,
		hs.HostID,
		hs.ServiceID,
		hs.Active,
		hs.ScheduleNumber,
		hs.ScheduleUnit,
		hs.LastCheck,
		hs.Status,
		time.Now(),
		hs.LastMessage,
		hs.ID,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// GetHostServiceStatus gets the status of a host service
func (m *postgresDBRepo) GetAllServiceStatusCount() (int, int, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	select	
	 (select count(id) from host_services where active = 1 and status = 'pending') as pending,
	 (select count(id) from host_services where active = 1 and status = 'healthy') as healthy,
	 (select count(id) from host_services where active = 1 and status = 'warning') as warning,
	 (select count(id) from host_services where active = 1 and status = 'problem') as problem
	`
	var pending, healthy, warning, problem int

	row := m.DB.QueryRowContext(ctx, query)
	err := row.Scan(&pending, &healthy, &warning, &problem)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return pending, healthy, warning, problem, nil
}

// GetServicesByStatus gets all services by status
func (m *postgresDBRepo) GetServicesByStatus(status string) ([]models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		select
			hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit, hs.last_check, hs.status, hs.created_at, hs.updated_at, h.host_name, s.service_name, hs.last_message
		from
			host_services hs
			left join hosts h on (hs.host_id = h.id)
			left join services s on (hs.service_id = s.id)
		where
			status = $1
			and hs.active = 1
		order by
			h.host_name, s.service_name
	`
	var hostServices []models.HostService

	rows, err := m.DB.QueryContext(ctx, query, status)
	if err != nil {
		return hostServices, err
	}
	defer rows.Close()

	for rows.Next() {
		var h models.HostService
		err = rows.Scan(
			&h.ID,
			&h.HostID,
			&h.ServiceID,
			&h.Active,
			&h.ScheduleNumber,
			&h.ScheduleUnit,
			&h.LastCheck,
			&h.Status,
			&h.CreatedAt,
			&h.UpdatedAt,
			&h.HostName,
			&h.Service.ServiceName,
			&h.LastMessage,
		)
		if err != nil {
			return nil, err
		}
		hostServices = append(hostServices, h)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return hostServices, nil
}

// GetHostServiceByID gets a host service by id
func (m *postgresDBRepo) GetHostServiceByID(id int) (models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		select 
			hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit, hs.last_check, hs.status, hs.created_at, hs.updated_at, s.id, s.service_name, s.icon, s.created_at, s.updated_at, h.host_name, hs.last_message
		from host_services hs 
			left join services s on (hs.service_id = s.id)
			left join hosts h on (hs.host_id = h.id)
		where 
			hs.id = $1`

	var hs models.HostService

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&hs.ID,
		&hs.HostID,
		&hs.ServiceID,
		&hs.Active,
		&hs.ScheduleNumber,
		&hs.ScheduleUnit,
		&hs.LastCheck,
		&hs.Status,
		&hs.CreatedAt,
		&hs.UpdatedAt,
		&hs.Service.ID,
		&hs.Service.ServiceName,
		&hs.Service.Icon,
		&hs.Service.CreatedAt,
		&hs.Service.UpdatedAt,
		&hs.HostName,
		&hs.LastMessage,
	)

	if err != nil {
		log.Println(err)
		return hs, err
	}

	return hs, nil
}

// GetServicesToMonitor gets all services to monitor
func (m *postgresDBRepo) GetServicesToMonitor() ([]models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		select
			hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit, hs.last_check, hs.status, hs.created_at, hs.updated_at, s.id, s.service_name, s.active, s.icon, s.created_at, s.updated_at, h.host_name, hs.last_message
		from
			host_services hs
			left join services s on (hs.service_id = s.id)
			left join hosts h on (hs.host_id = h.id)
		where
			h.active = 1
			and hs.active = 1
	`

	var services []models.HostService

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		log.Println(err)
		return services, err
	}
	defer rows.Close()

	for rows.Next() {
		var h models.HostService
		err = rows.Scan(
			&h.ID,
			&h.HostID,
			&h.ServiceID,
			&h.Active,
			&h.ScheduleNumber,
			&h.ScheduleUnit,
			&h.LastCheck,
			&h.Status,
			&h.CreatedAt,
			&h.UpdatedAt,
			&h.Service.ID,
			&h.Service.ServiceName,
			&h.Service.Active,
			&h.Service.Icon,
			&h.Service.CreatedAt,
			&h.Service.UpdatedAt,
			&h.HostName,
			&h.LastMessage,
		)
		if err != nil {
			log.Println(err)
			return services, err
		}

		services = append(services, h)
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return services, err
	}

	return services, nil
}

// GetHostServiceByHostIDServiceID gets a host service by host id and service id
func (m *postgresDBRepo) GetHostServiceByHostIDServiceID(hostID, serviceID int) (models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	select 
		hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit, hs.last_check, hs.status, hs.created_at, hs.updated_at, s.id, s.service_name, s.icon, s.created_at, s.updated_at, h.host_name, hs.last_message 
	from 
		host_services hs 
		left join services s on (hs.service_id = s.id) 
		left join hosts h on (hs.host_id = h.id) 
	where 
		hs.host_id = $1 and hs.service_id = $2`

	var hs models.HostService

	row := m.DB.QueryRowContext(ctx, query, hostID, serviceID)

	err := row.Scan(
		&hs.ID,
		&hs.HostID,
		&hs.ServiceID,
		&hs.Active,
		&hs.ScheduleNumber,
		&hs.ScheduleUnit,
		&hs.LastCheck,
		&hs.Status,
		&hs.CreatedAt,
		&hs.UpdatedAt,
		&hs.Service.ID,
		&hs.Service.ServiceName,
		&hs.Service.Icon,
		&hs.Service.CreatedAt,
		&hs.Service.UpdatedAt,
		&hs.HostName,
		&hs.LastMessage,
	)

	if err != nil {
		log.Println(err)
		return hs, err
	}

	return hs, nil
}

// InsertEvent inserts an event into the database
func (m *postgresDBRepo) InsertEvent(e models.Events) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		insert into events (host_service_id, event_type, host_id, service_name, host_name, message, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := m.DB.ExecContext(ctx, stmt,
		e.HostServiceID,
		e.EventType,
		e.HostID,
		e.ServiceName,
		e.HostName,
		e.Message,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// GetAllEvents gets all events
func (m *postgresDBRepo) GetAllEvents() ([]models.Events, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		select
			id, host_service_id, event_type, host_id, service_name, host_name, message, created_at, updated_at
		from
			events 
		order by
			created_at desc
	`

	var events []models.Events

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		log.Println(err)
		return events, err
	}

	defer rows.Close()

	for rows.Next() {
		var e models.Events
		err = rows.Scan(
			&e.ID,
			&e.HostServiceID,
			&e.EventType,
			&e.HostID,
			&e.ServiceName,
			&e.HostName,
			&e.Message,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			log.Println(err)
			return events, err
		}
		events = append(events, e)
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return events, err
	}

	return events, nil
}
