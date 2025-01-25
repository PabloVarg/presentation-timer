-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE clean_section_positions()
LANGUAGE plpgsql
AS $$
DECLARE
    p RECORD;
BEGIN
    FOR p IN SELECT id FROM presentation FOR UPDATE LOOP
        WITH ordered AS (
            SELECT
                id,
                ROW_NUMBER() OVER (ORDER BY position) AS new_position
            FROM section
            WHERE presentation = p.id
        )
        UPDATE section
        SET position = ordered.new_position
        FROM ordered
        WHERE section.id = ordered.id;
    END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP PROCEDURE clean_section_positions()
-- +goose StatementEnd
