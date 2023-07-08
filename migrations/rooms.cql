USE clique;

CREATE TABLE IF NOT EXISTS rooms (
    id UUID,
    room_name TEXT,
    channels SET<UUID>,
    members MAP<UUID, FROZEN<SET<string>>>,
    created_at TIMESTAMP,
    admin UUID,
    PRIMARY KEY((id, admin), created_at)
) WITH CLUSTERING ORDER BY (created_at DESC);


CREATE TABLE IF NOT EXISTS channels (
    id UUID,
    channel_name TEXT,
    room_id UUID,
    members SET<string>,
    is_direct_channel BOOLEAN,
    created_at TIMESTAMP,
    PRIMARY KEY((room_id, id), created_at)
) WITH CLUSTERING ORDER BY (created_at DESC);


CREATE TABLE IF NOT EXISTS messages (
    channel_id UUID,
    sender string,
    room_id UUID, 
    content TEXT,
    timestamp TIMESTAMP,
    type TEXT,
    PRIMARY KEY((channel_id), timestamp)
) WITH CLUSTERING ORDER BY (timestamp DESC);


INSERT INTO rooms (id, room_name, channels, members, created_at, admin)
VALUES (00000000-0000-0000-0000-000000000000, '', null, null, toTimestamp(now()), 00000000-0000-0000-0000-000000000000);

