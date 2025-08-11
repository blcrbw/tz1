package subscription

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"tz1/internal/subscription"
	"tz1/pkg/client/postgresql"
	"tz1/pkg/helper"
	"tz1/pkg/logging"
)

type repository struct {
	client postgresql.Client
	logger *logging.Logger
}

type pgSubscription struct {
	s       *subscription.Subscription
	pgStart pgtype.Date
	pgEnd   pgtype.Date
}

func (pgs *pgSubscription) Validate() error {
	var err error
	if pgs.s.StartDate != "" {
		pgs.pgStart, err = helper.ParsePgDate(pgs.s.StartDate)
		if err != nil {
			return err
		}
	}
	if pgs.s.EndDate != "" {
		pgs.pgEnd, err = helper.ParsePgDate(pgs.s.EndDate)
		if err != nil {
			return err
		}
	}
	if pgs.s.User != "" && !helper.IsValidUUID(pgs.s.User) {
		err = fmt.Errorf("invalid subscription User: %s", pgs.s.User)
		return err
	}
	if pgs.s.ID != "" && !helper.IsValidUUID(pgs.s.ID) {
		err = fmt.Errorf("invalid subscription ID: %s", pgs.s.ID)
		return err
	}
	if pgs.pgStart.Valid && pgs.pgEnd.Valid && pgs.pgEnd.Time.Before(pgs.pgStart.Time) {
		err = fmt.Errorf("end date (%s) cannot be earlier than start (%s)", pgs.pgEnd.Time.Format("01-2006"), pgs.pgStart.Time.Format("01-2006"))
		return err
	}
	return nil
}

func (r *repository) Create(ctx context.Context, s *subscription.Subscription) error {
	pgSubscription := pgSubscription{s: s}
	if err := pgSubscription.Validate(); err != nil {
		r.logger.Error(err)
		return err
	}

	q := `
		INSERT INTO public.subscription 
		    (service_name, price, "user", start_date, end_date ) 
		VALUES 
		       ($1, $2, $3, $4, $5) 
		RETURNING id
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", helper.FormatQuery(q)))
	row := r.client.QueryRow(ctx, q, pgSubscription.s.ServiceName, pgSubscription.s.Price, pgSubscription.s.User, pgSubscription.pgStart, pgSubscription.pgEnd)
	if err := row.Scan(&pgSubscription.s.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
			r.logger.Error(newErr)
			return newErr
		}
		return err
	}

	return nil
}

func (r *repository) GetList(ctx context.Context, limit int, offset int, from string, to string, user string, service string) (a []subscription.Subscription, err error) {

	whereSet := "WHERE "
	fromDate, err := helper.ParsePgDate(from)
	toDate, err := helper.ParsePgDate(to)
	placeholder := 1
	args := make([]interface{}, 0)

	q := `
		SELECT id, "user", service_name, price, to_char(start_date, 'MM-YYYY'), to_char(end_date, 'MM-YYYY')  
		FROM public.subscription
	`
	if fromDate.Valid && toDate.Valid && fromDate.Time.Before(toDate.Time) {
		q = fmt.Sprintf("%s \n\t\t%s(start_date between $%d AND $%d)", q, whereSet, placeholder, placeholder+1)
		whereSet = "AND"
		args = append(args, fromDate)
		args = append(args, toDate)
		placeholder += 2
	} else if fromDate.Valid && toDate.Valid && fromDate.Time.Equal(toDate.Time) {
		q = fmt.Sprintf("%s \n\t\t%s start_date = $%d", q, whereSet, placeholder)
		args = append(args, toDate)
		whereSet = "AND"
		placeholder++
	} else if fromDate.Valid && toDate.Valid && fromDate.Time.After(toDate.Time) {
		err = fmt.Errorf("end date (%s) cannot be earlier than start (%s)", toDate.Time.Format("01-2006"), fromDate.Time.Format("01-2006"))
		return nil, err
	} else if fromDate.Valid && !toDate.Valid {
		q = fmt.Sprintf("%s \n\t\t%s start_date >= $%d", q, whereSet, placeholder)
		args = append(args, fromDate)
		whereSet = "AND"
		placeholder++
	} else if !fromDate.Valid && toDate.Valid {
		q = fmt.Sprintf("%s \n\t\t%s start_date <= $%d", q, whereSet, placeholder)
		args = append(args, toDate)
		whereSet = "AND"
		placeholder++
	}

	if user != "" {
		if helper.IsValidUUID(user) {
			q = fmt.Sprintf("%s %s \"user\" = $%d", q, whereSet, placeholder)
			args = append(args, user)
			whereSet = "AND"
			placeholder++
		} else {
			return nil, fmt.Errorf("invalid subscription User: %s", user)
		}
	}
	if service != "" {
		q = fmt.Sprintf("%s %s service_name = $%d", q, whereSet, placeholder)
		args = append(args, service)
		whereSet = "AND"
		placeholder++
	}
	q = fmt.Sprintf("%s \n\t\tORDER BY start_date ASC", q)
	q = fmt.Sprintf("%s \n\t\tLIMIT $%d OFFSET $%d;", q, placeholder, placeholder+1)
	placeholder += 2
	args = append(args, limit, offset)

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", helper.FormatQuery(q)))

	rows, err := r.client.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	subscriptions := make([]subscription.Subscription, 0)

	for rows.Next() {
		var s subscription.Subscription

		var nullableEndDate pgtype.Text

		err = rows.Scan(&s.ID, &s.User, &s.ServiceName, &s.Price, &s.StartDate, &nullableEndDate)
		if err != nil {
			return nil, err
		}

		if nullableEndDate.Valid {
			s.EndDate = nullableEndDate.String
		}

		subscriptions = append(subscriptions, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (r *repository) GetSum(ctx context.Context, from string, to string, user string, service string) (sum int64, err error) {
	var nullableInt pgtype.Int8

	fromDate, err := helper.ParsePgDate(from)
	toDate, err := helper.ParsePgDate(to)

	placeholder := 1
	args := make([]interface{}, 0)

	q := `
		SELECT SUM(price)  
		FROM public.subscription
	`
	if fromDate.Valid && toDate.Valid && fromDate.Time.Before(toDate.Time) {
		q = fmt.Sprintf("%s \n\t\tWHERE (start_date between $1 AND $2)", q)
		args = append(args, fromDate)
		args = append(args, toDate)
		placeholder = 3
	} else if fromDate.Valid && toDate.Valid && fromDate.Time.Equal(toDate.Time) {
		q = fmt.Sprintf("%s \n\t\tWHERE start_date = $%d", q, placeholder)
		args = append(args, toDate)
		placeholder++
	} else if fromDate.Valid && toDate.Valid && fromDate.Time.After(toDate.Time) {
		err = fmt.Errorf("end date (%s) cannot be earlier than start (%s)", toDate.Time.Format("01-2006"), fromDate.Time.Format("01-2006"))
		return 0, err
	} else if fromDate.Valid && !toDate.Valid {
		q = fmt.Sprintf("%s \n\t\tWHERE start_date >= $%d", q, placeholder)
		args = append(args, fromDate)
		placeholder++
	} else if !fromDate.Valid && toDate.Valid {
		q = fmt.Sprintf("%s \n\t\tWHERE start_date <= $%d", q, placeholder)
		args = append(args, toDate)
		placeholder++
	} else if !fromDate.Valid && !toDate.Valid {
		err = fmt.Errorf("date range is not specified")
		return 0, err
	}

	if user != "" {
		if helper.IsValidUUID(user) {
			q = fmt.Sprintf("%s AND \"user\" = $%d", q, placeholder)
			args = append(args, user)
			placeholder++
		} else {
			return sum, fmt.Errorf("invalid subscription User: %s", user)
		}
	}
	if service != "" {
		q = fmt.Sprintf("%s AND service_name = $%d", q, placeholder)
		args = append(args, service)
		placeholder++
	}
	q = q + ";"
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", helper.FormatQuery(q)))

	row := r.client.QueryRow(ctx, q, args...)
	if err := row.Scan(&nullableInt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
			r.logger.Error(newErr)
			return 0, newErr
		}
		return 0, err
	}

	return nullableInt.Int64, nil
}

func (r *repository) FindAll(ctx context.Context) (a []subscription.Subscription, err error) {
	q := `
		SELECT id, "user", service_name, price, to_char(start_date, 'MM-YYYY'), to_char(end_date, 'MM-YYYY')  
		FROM public.subscription
	`
	q = q + ";"
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", helper.FormatQuery(q)))

	rows, err := r.client.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	subscriptions := make([]subscription.Subscription, 0)

	for rows.Next() {
		var s subscription.Subscription

		var nullableEndDate pgtype.Text

		err = rows.Scan(&s.ID, &s.User, &s.ServiceName, &s.Price, &s.StartDate, &nullableEndDate)
		if err != nil {
			return nil, err
		}

		if nullableEndDate.Valid {
			s.EndDate = nullableEndDate.String
		}

		subscriptions = append(subscriptions, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (r *repository) FindOne(ctx context.Context, id string) (subscription.Subscription, error) {
	q := `
		SELECT id, "user", service_name, price, to_char(start_date, 'MM-YYYY'), to_char(end_date, 'MM-YYYY')  
		FROM public.subscription 
		WHERE id = $1
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", helper.FormatQuery(q)))

	var s subscription.Subscription
	var nullableEndDate pgtype.Text
	row := r.client.QueryRow(ctx, q, id)
	err := row.Scan(&s.ID, &s.User, &s.ServiceName, &s.Price, &s.StartDate, &nullableEndDate)
	if err != nil {
		return subscription.Subscription{}, err
	}

	if nullableEndDate.Valid {
		s.EndDate = nullableEndDate.String
	}

	return s, nil
}

func (r *repository) Update(ctx context.Context, id string, s *subscription.Subscription) error {
	s.ID = id
	pgSubscription := pgSubscription{s: s}
	if err := pgSubscription.Validate(); err != nil {
		r.logger.Error(err)
		return err
	}

	q := `
		UPDATE public.subscription 
		SET service_name = $1,
		    price = $2,
		    "user" = $3,
		    start_date = $4,
		    end_date = $5
		WHERE id = $6 
		RETURNING id
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", helper.FormatQuery(q)))

	row := r.client.QueryRow(ctx, q, pgSubscription.s.ServiceName, pgSubscription.s.Price, pgSubscription.s.User, pgSubscription.pgStart, pgSubscription.pgEnd, pgSubscription.s.ID)

	if err := row.Scan(&pgSubscription.s.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
			r.logger.Error(newErr)
			return newErr
		}
		return err
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	q := `
		DELETE FROM public.subscription 
	    WHERE id = $1
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", helper.FormatQuery(q)))

	_, err := r.client.Query(ctx, q, id)

	return err
}

func NewRepository(client postgresql.Client, logger *logging.Logger) subscription.Repository {
	return &repository{
		client: client,
		logger: logger,
	}
}
