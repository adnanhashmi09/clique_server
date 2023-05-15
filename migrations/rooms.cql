USE clique;

CREATE TABLE IF NOT EXISTS rooms (
    id UUID,
    room_name TEXT,
    channels SET<UUID>,
    members MAP<UUID, FROZEN<SET<UUID>>>,
    created_at TIMESTAMP,
    admin UUID,
    PRIMARY KEY((id, admin), created_at)
) WITH CLUSTERING ORDER BY (created_at DESC);


CREATE TABLE IF NOT EXISTS channels (
    id UUID,
    channel_name TEXT,
    room_id UUID,
    members SET<UUID>,
    is_direct_channel BOOLEAN,
    messages LIST<INT>,
    created_at TIMESTAMP,
    PRIMARY KEY((room_id, id), created_at)
) WITH CLUSTERING ORDER BY (created_at DESC);


CREATE TABLE IF NOT EXISTS messages (
    id INT,
    channel_id UUID,
    sender_id UUID,
    content TEXT,
    timestamp TIMESTAMP,
    PRIMARY KEY((channel_id, sender_id, id), timestamp)
) WITH CLUSTERING ORDER BY (timestamp DESC);
