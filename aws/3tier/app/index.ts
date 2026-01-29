import { serve } from "bun";
import { Pool } from "pg";

const pool = new Pool({
  host: process.env.DB_HOST,
  port: 5432,
  database: "postgres",
  user: "postgres",
  password: process.env.DB_PASSWORD,
});

serve({
  port: 80,
  async fetch(req) {
    const url = new URL(req.url);

    if (url.pathname === "/health") {
      return new Response("OK");
    }

    if (url.pathname === "/db") {
      try {
        const result = await pool.query(`
          SELECT
            current_database() as database,
            current_user as user,
            version() as version,
            inet_server_addr() as server_ip,
            pg_postmaster_start_time() as uptime_since,
            (SELECT count(*) FROM pg_stat_activity) as active_connections
        `);
        return Response.json({
          ...result.rows[0],
          app_host: process.env.HOSTNAME
        });
      } catch (err) {
        return Response.json({ error: String(err) }, { status: 500 });
      }
    }

    return Response.json({
      message: "Hello from Bun!",
      host: process.env.HOSTNAME,
    });
  },
});
