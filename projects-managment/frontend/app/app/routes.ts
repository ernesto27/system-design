import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  route("login", "routes/login.tsx"),
  route("admin/create-project", "routes/admin.create-project.tsx"),
  route("admin/projects", "routes/projects.list.tsx"), // Add the new route
] satisfies RouteConfig;
