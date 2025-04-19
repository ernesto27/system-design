// Define shared TypeScript interfaces here

export interface ProjectStatus {
  id: number;
  name: string;
}

export interface Role {
  id: number;
  name: string;
  percentage: number;
  _uniqueId?: number; // Optional unique identifier for UI purposes
}

export interface Project {
  name: string;
  description: string;
  projectStatusId: number;
  timeEstimation?: number;
  Roles?: Role[];
}

