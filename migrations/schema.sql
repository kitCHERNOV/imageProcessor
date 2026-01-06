CREATE TABLE images (
    id INTEGER PRIMARY KEY,
    original_filename TEXT,
    original_path TEXT,
    mime_type TEXT,
    file_size INTEGER,
    status TEXT DEFAULT 'pending',
    action TEXT, -- type of action: ["resize", "miniature", "watermark"]
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);