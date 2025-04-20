import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  route("login", "routes/login.tsx"),
  route("admin/create-project", "routes/admin.create-project.tsx"),
  route("admin/projects", "routes/projects.list.tsx"),
  route("admin/projects/:id", "routes/project.detail.tsx"), // Add project detail route with ID parameter
] satisfies RouteConfig;
