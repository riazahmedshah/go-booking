package property

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/riazahmedshah/go-booking/internal/errs"
	"github.com/riazahmedshah/go-booking/internal/server"
)

type PropertyRepository struct {
	server *server.Server
}

func NewPropertyRepository(server *server.Server) *PropertyRepository {
	return &PropertyRepository{server: server}
}

func (pr *PropertyRepository) Createproperty(ctx context.Context, tx pgx.Tx, hostID string, payload *CreatePropertyPayload) (*Property, error) {
	stmt := `
		INSERT INTO properties(
			host_id, title, sub_title, max_guests, price
		)
		VALUES (
			@host_id, @title, @sub_title, @max_guests, @price
		)
		RETURNING *
	`

	rows, err := tx.Query(ctx, stmt, pgx.NamedArgs{
		"host_id":    hostID,
		"title":      payload.Title,
		"sub_title":  payload.SubTitle,
		"max_guests": payload.MaxGuests,
		"price":      payload.Price,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique key violation
			return nil, errs.ErrPropertyTitleExists
		}

		return nil, fmt.Errorf("failed to execute create property query for host_id=%s title=%s: %w", hostID, payload.Title, err)
	}

	propertyItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Property])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:properties for host_id=%s title=%s: %w", hostID, payload.Title, err)
	}

	return &propertyItem, nil
}

func (pr *PropertyRepository) CreateAddress(ctx context.Context, tx pgx.Tx, payload *CreateAddressPayload) (*Address, error) {
	stmt := `
		INSERT INTO addresses (country, state, pincode, city, area, property_id)
		VALUES (@country, @state, @pincode, @city, @area, @property_id)
		RETURNING *
	`

	args := pgx.NamedArgs{
		"country":     payload.Country,
		"state":       payload.State,
		"pincode":     payload.Pincode,
		"city":        payload.City,
		"area":        payload.Area,
		"property_id": payload.PropertyID,
	}

	rows, err := tx.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute create address query: %w", err)
	}
	defer rows.Close()

	address, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Address])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:addresses: %w", err)
	}

	return &address, nil
}

func (pr *PropertyRepository) CreatePropertyImages(ctx context.Context, tx pgx.Tx, propertyID string, keys []string) ([]*PropertyImages, error) {
	stmt := `
		INSERT INTO property_images (property_id, key)
		SELECT @property_id, unnest(@key::text[])
		RETURNING *
	`

	rows, err := tx.Query(ctx, stmt, pgx.NamedArgs{
		"property_id": propertyID,
		"key":         keys,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute create property images query: %w", err)
	}
	defer rows.Close()

	images, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[PropertyImages])
	if err != nil {
		return nil, fmt.Errorf("failed to collect rows from table:property_images: %w", err)
	}

	return images, nil
}

func (pr *PropertyRepository) UpdateImageStatus(ctx context.Context, imageID string, status string) error {
	stmt := `
		UPDATE property_images
		SET status = @status
		WHERE id = @id
	`

	result, err := pr.server.DB.Exec(ctx, stmt, pgx.NamedArgs{
		"id":     imageID,
		"status": status,
	})
	if err != nil {
		return fmt.Errorf("failed to execute update image status query: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errs.ErrImageNotFound
	}

	return nil
}

func (pr *PropertyRepository) GetAllProperties(ctx context.Context) ([]*PopulatedProperty, error) {
	stmt := `
		SELECT
			p.id, p.title, p.sub_title, p.price, p.host_id, p.max_guests, 
			CASE 
				WHEN a.id IS NOT NULL THEN jsonb_build_object(
					'id', a.id,
					'country', a.country,
					'state', a.state,
					'pincode', a.pincode,
					'city', a.city,
					'area', a.area,
					'propertyId', a.property_id
				)
				ELSE NULL
		  END AS address,
			COALESCE(
			jsonb_agg(
				jsonb_build_object(
					'id', pi.id,
					'key', pi.key,
					'status', pi.status
				)
			) FILTER (WHERE pi.id IS NOT NULL),
			'[]'::JSONB
			) AS images,
			p.created_at, p.updated_at
		FROM properties p
		INNER JOIN addresses a ON p.id = a.property_id
		LEFT JOIN property_images pi ON p.id = pi.property_id
		GROUP BY p.id, a.id
		ORDER BY p.created_at DESC
	`
	rows, err := pr.server.DB.Query(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get all properties query: %w", err)
	}
	defer rows.Close()

	properties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[PopulatedProperty])
	if err != nil {
		return nil, fmt.Errorf("failed to collect rows from table:properties: %w", err)
	}

	return properties, nil
}

func (pr *PropertyRepository) GetPropertyByID(ctx context.Context, propertyID string) (*PopulatedPropertyWithHost, error) {
	stmt := `
		SELECT
			p.id, p.title, p.sub_title, p.price, p.host_id, p.max_guests, 
			CASE
				WHEN u.id IS NOT NULL THEN jsonb_build_object(
					'id', u.id,
					'name', u.first_name
				)
				ELSE NULL
			END AS host,
			CASE 
				WHEN a.id IS NOT NULL THEN jsonb_build_object(
					'id', a.id,
					'country', a.country,
					'state', a.state,
					'pincode', a.pincode,
					'city', a.city,
					'area', a.area,
					'propertyId', a.property_id
				)
				ELSE NULL
		  END AS address,
			COALESCE(
			jsonb_agg(
				jsonb_build_object(
					'id', pi.id,
					'key', pi.key,
					'status', pi.status
				)
			) FILTER (WHERE pi.id IS NOT NULL),
			'[]'::JSONB
			) AS images,
			p.created_at, p.updated_at
		FROM properties p
		INNER JOIN users u ON p.host_id = u.id
		INNER JOIN addresses a ON p.id = a.property_id
		LEFT JOIN property_images pi ON p.id = pi.property_id
		WHERE p.id = @id
		GROUP BY p.id, u.id, a.id
		ORDER BY p.created_at DESC
	`

	rows, err := pr.server.DB.Query(ctx, stmt, pgx.NamedArgs{
		"id": propertyID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute get property by id for property_id %v: %w", propertyID, err)
	}

	propertyItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[PopulatedPropertyWithHost])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrPropertyNotFound
		}
		return nil, fmt.Errorf("failed to collect row from table:properties for property_id=%s: %w", propertyID, err)
	}

	return &propertyItem, nil

}

func (pr *PropertyRepository) UpdateProperty(ctx context.Context, propertyID string, payload *UpdatePropertyPayload) (*Property, error) {
	stmt := "UPDATE properties SET "

	args := pgx.NamedArgs{
		"id": propertyID,
	}

	setClauses := []string{}

	if payload.SubTitle != nil {
		setClauses = append(setClauses, "sub_title = @sub_title")
		args["sub_title"] = *payload.SubTitle
	}

	if payload.AddressID != nil {
		setClauses = append(setClauses, "address_id = @address_id")
		args["address_id"] = *payload.AddressID
	}

	if payload.MaxGuests != nil {
		setClauses = append(setClauses, "max_guests = @max_guests")
		args["max_guests"] = *payload.MaxGuests
	}

	if len(setClauses) == 0 {
		return nil, errs.ErrBadUpdateRequest
	}

	stmt += strings.Join(setClauses, ", ")
	stmt += " WHERE id = @id RETURNING *"

	rows, err := pr.server.DB.Query(ctx, stmt, args)

	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	updatedProperty, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Property])

	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:properties for property_id=%s: %w", propertyID, err)
	}

	return &updatedProperty, nil

}

func (pr *PropertyRepository) DeleteProperty(ctx context.Context, propertyID string) error {
	stmt := `
		DELETE FROM properties
		WHERE id = @id
	`

	result, err := pr.server.DB.Exec(ctx, stmt, pgx.NamedArgs{
		"id": propertyID,
	})

	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errs.ErrPropertyNotFound
	}

	return nil
}

func (pr *PropertyRepository) GetPropertyAvailability(ctx context.Context, propertyID string) ([]*PropertyAvailabiliy, error) {
	stmt := `
		SELECT
			id, property_id, date, is_available, booking_id, created_at, updated_at
		FROM
			property_availability
		WHERE
			property_id = @property_id
	`
	rows, err := pr.server.DB.Query(ctx, stmt, pgx.NamedArgs{
		"property_id": propertyID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	availability, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[PropertyAvailabiliy])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:property_availability for property_id=%s: %w", propertyID, err)
	}

	return availability, nil
}

func (pr *PropertyRepository) GetHostListings(ctx context.Context, hostID string) ([]*PopulatedProperty, error) {
	// stmt := `
	// 	SELECT p.id, p.title, p.sub_title AS "subTitle", p.price, p.host_id AS "hostId", p.max_guests AS "maxGuests", p.created_at AS "createdAt", p.updated_at AS "updatedAt",
	// 	COALESCE(
	// 		json_agg(
	// 			json_build_object(
	// 				'id', pi.id,
	// 				'key', pi.key,
	// 				'status', pi.status
	// 			)
	// 		) FILTER (WHERE pi.id IS NOT NULL),
	// 		'[]'
	// 	) AS images
	// 	FROM properties p
	// 	LEFT JOIN property_images pi ON p.id = pi.property_id
	// 	WHERE p.host_id = @host_id
	// 	GROUP BY p.id
	// 	ORDER BY p.created_at DESC
	// `

	stmt := `
		SELECT
			p.id, p.title, p.sub_title, p.price, p.host_id, p.max_guests, 
			CASE 
				WHEN a.id IS NOT NULL THEN jsonb_build_object(
					'id', a.id,
					'country', a.country,
					'state', a.state,
					'pincode', a.pincode,
					'city', a.city,
					'area', a.area,
					'propertyId', a.property_id
				)
				ELSE NULL
		  END AS address,
			COALESCE(
			jsonb_agg(
				jsonb_build_object(
					'id', pi.id,
					'key', pi.key,
					'status', pi.status
				)
			) FILTER (WHERE pi.id IS NOT NULL),
			'[]'::JSONB
			) AS images,
			p.created_at, p.updated_at
		FROM properties p
		INNER JOIN addresses a ON p.id = a.property_id
		LEFT JOIN property_images pi ON p.id = pi.property_id
		WHERE p.host_id = @host_id
		GROUP BY p.id, a.id
		ORDER BY p.created_at DESC
	`

	rows, err := pr.server.DB.Query(ctx, stmt, pgx.NamedArgs{
		"host_id": hostID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	listings, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[PopulatedProperty])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:properties for host_id=%s: %w", hostID, err)
	}

	return listings, nil
}
