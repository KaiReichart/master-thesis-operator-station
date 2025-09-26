-- aircraft_type definition

CREATE TABLE aircraft_type (
    type text primary key,
    category text,
    wing_span integer,
    engine_type integer,
    nof_engines integer
);

CREATE INDEX aircraft_type_idx on aircraft_type (type collate nocase);


-- enum_backup_period definition

CREATE TABLE "enum_backup_period" (
    id integer primary key,
    sym_id text not null,
    name text,
    desc text
);

CREATE UNIQUE INDEX enum_backup_period_idx1 on enum_backup_period (sym_id);


-- enum_country definition

CREATE TABLE enum_country(
    id integer primary key,
    sym_id text not null,
    name text,
    desc text
);

CREATE UNIQUE INDEX enum_country_idx1 on enum_country(sym_id);


-- enum_engine_event definition

CREATE TABLE enum_engine_event(
    id integer primary key,
    sym_id text not null,
    name text,
    desc text
);

CREATE UNIQUE INDEX enum_engine_event_idx1 on enum_engine_event(sym_id);


-- enum_location_category definition

CREATE TABLE enum_location_category(
    id integer primary key,
    sym_id text not null,
    name text,
    desc text
);

CREATE UNIQUE INDEX enum_location_category_idx1 on enum_location_category(sym_id);


-- enum_location_type definition

CREATE TABLE enum_location_type(
    id integer primary key,
    sym_id text not null,
    name text,
    desc text
);

CREATE UNIQUE INDEX enum_location_type_idx1 on enum_location_type(sym_id);


-- flight definition

CREATE TABLE "flight" (
    id integer primary key,
    creation_time datetime not null default (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    user_aircraft_seq_nr integer not null,
    title text,
    description text,
    flight_number text,
    surface_type integer,
    surface_condition integer,
    on_any_runway integer,
    on_parking_spot integer,
    ground_altitude real,
    ambient_temperature real,
    total_air_temperature real,
    wind_speed real,
    wind_direction real,
    visibility real,
    sea_level_pressure real,
    pitot_icing real,
    structural_icing real,
    precipitation_state integer,
    in_clouds integer,
    start_local_sim_time datetime not null default (strftime('%Y-%m-%dT%H:%M:%f', 'now', 'localtime')),
    start_zulu_sim_time datetime not null default (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    end_local_sim_time datetime not null default (strftime('%Y-%m-%dT%H:%M:%f', 'now', 'localtime')),
    end_zulu_sim_time datetime not null default (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);


-- migr definition

CREATE TABLE migr(id text not null,step integer not null,success integer not null,timestamp datetime default current_timestamp,msg text,primary key (id, step));


-- aircraft definition

CREATE TABLE "aircraft" (
    id integer primary key,
    flight_id integer not null,
    seq_nr integer not null,
    type text not null,
    time_offset integer,
    tail_number text,
    airline text,
    initial_airspeed integer,
    altitude_above_ground real,
    start_on_ground integer,
    foreign key(flight_id) references flight(id)
    foreign key(type) references aircraft_type(type)
);

CREATE UNIQUE INDEX aircraft_idx1 on aircraft (flight_id, seq_nr);
CREATE INDEX aircraft_idx2 on aircraft (type collate nocase);


-- attitude definition

CREATE TABLE attitude (
    aircraft_id integer not null,
    timestamp integer not null,
    pitch real,
    bank real,
    true_heading real,
    velocity_x real,
    velocity_y real,
    velocity_z real,
    on_ground int,
    primary key(aircraft_id, timestamp),
    foreign key(aircraft_id) references aircraft(id)
);


-- engine definition

CREATE TABLE "engine" (
    aircraft_id integer not null,
    timestamp integer not null,
    throttle_lever_position1 real,
    throttle_lever_position2 real,
    throttle_lever_position3 real,
    throttle_lever_position4 real,
    propeller_lever_position1 real,
    propeller_lever_position2 real,
    propeller_lever_position3 real,
    propeller_lever_position4 real,
    mixture_lever_position1 real,
    mixture_lever_position2 real,
    mixture_lever_position3 real,
    mixture_lever_position4 real,
    cowl_flap_position1 real,
    cowl_flap_position2 real,
    cowl_flap_position3 real,
    cowl_flap_position4 real,
    electrical_master_battery1 integer,
    electrical_master_battery2 integer,
    electrical_master_battery3 integer,
    electrical_master_battery4 integer,
    general_engine_starter1 integer,
    general_engine_starter2 integer,
    general_engine_starter3 integer,
    general_engine_starter4 integer, general_engine_combustion1 integer, general_engine_combustion2 integer, general_engine_combustion3 integer, general_engine_combustion4 integer,
    primary key(aircraft_id, timestamp),
    foreign key(aircraft_id) references aircraft(id)
);


-- handle definition

CREATE TABLE handle (
    aircraft_id integer not null,
    timestamp integer not null,
    brake_left_position integer,
    brake_right_position integer,
    water_rudder_handle_position integer,
    tailhook_position integer,
    canopy_open integer,
    left_wing_folding integer,
    right_wing_folding integer,
    gear_handle_position integer, tailhook_handle_position integer, folding_wing_handle_position integer, steer_input_control real,
    primary key(aircraft_id, timestamp),
    foreign key(aircraft_id) references aircraft(id)
);


-- light definition

CREATE TABLE light (
    aircraft_id integer not null,
    timestamp integer not null,
    light_states integer,
    primary key(aircraft_id, timestamp),
    foreign key(aircraft_id) references aircraft(id)
);


-- location definition

CREATE TABLE location(
    id integer primary key,
    title text,
    description text,
    type_id integer,
    category_id integer,
    country_id integer,
    identifier text,
    latitude real,
    longitude real,
    altitude real,
    pitch real,
    bank real,
    true_heading real,
    indicated_airspeed integer,
    on_ground integer,
    attributes integer, engine_event integer references enum_engine_event(id),
    foreign key(type_id) references enum_location_type(id)
    foreign key(category_id) references enum_location_category(id)
    foreign key(country_id) references enum_country(id)
);

CREATE INDEX location_idx1 on location(title collate nocase);
CREATE INDEX location_idx2 on location(description collate nocase);
CREATE INDEX location_idx3 on location(category_id);
CREATE INDEX location_idx4 on location(country_id);
CREATE INDEX location_idx5 on location(identifier collate nocase);
CREATE INDEX location_idx6 on location(on_ground);


-- metadata definition

CREATE TABLE metadata (
    creation_date datetime,
    app_version text,
    last_optim_date datetime,
    last_backup_date datetime,
    backup_directory_path text,
    backup_period_id integer, next_backup_date datetime,
    foreign key(backup_period_id) references "enum_backup_period"(id)
);


-- "position" definition

CREATE TABLE "position" (
    aircraft_id integer not null,
    timestamp integer not null,
    latitude real,
    longitude real,
    altitude real,
    indicated_altitude real, calibrated_indicated_altitude real, pressure_altitude real,
    primary key(aircraft_id, timestamp),
    foreign key(aircraft_id) references aircraft(id)
);


-- primary_flight_control definition

CREATE TABLE primary_flight_control (
    aircraft_id integer not null,
    timestamp integer not null,
    rudder_position integer,
    elevator_position integer,
    aileron_position integer, rudder_deflection real, elevator_deflection real, aileron_left_deflection real, aileron_right_deflection real,
    primary key(aircraft_id, timestamp),
    foreign key(aircraft_id) references aircraft(id)
);


-- secondary_flight_control definition

CREATE TABLE secondary_flight_control (
    aircraft_id integer not null,
    timestamp integer not null,
    left_leading_edge_flaps_position integer,
    right_leading_edge_flaps_position integer,
    left_trailing_edge_flaps_position integer,
    right_trailing_edge_flaps_position integer,
    spoilers_handle_percent integer,
    flaps_handle_index integer, left_spoilers_position integer, right_spoilers_position integer, spoilers_armed integer,
    primary key(aircraft_id, timestamp),
    foreign key(aircraft_id) references aircraft(id)
);


-- waypoint definition

CREATE TABLE waypoint (
    aircraft_id integer not null,
    timestamp integer not null,
    ident text,
    latitude real,
    longitude real,
    altitude real,
    local_sim_time datetime,
    zulu_sim_time datetime,
    primary key(aircraft_id, timestamp),
    foreign key(aircraft_id) references aircraft(id)
);

CREATE INDEX waypoint_idx1 on waypoint (ident collate nocase);
