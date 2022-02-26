BEGIN;

CREATE TABLE login_boost_state
(
    char_id bigint REFERENCES characters (id),	
	week_req uint8,
	week_count uint8,
	available bool,
    end_time int,
	CONSTRAINT id_week UNIQUE(char_id, week_req)
);	

END;