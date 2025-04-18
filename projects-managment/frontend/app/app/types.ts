// Define shared TypeScript interfaces here

export interface ProjectStatus {
  id: number;
  name: string;
}

export interface Project {
  id: number;
  name: string;
  description: string;
  project_status_id: number;
  created_at: string; 
  updated_at: string;
}

