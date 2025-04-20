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

export interface Status {
  id: number;
  name: string;
}

export interface Project {
  id?: number;
  name: string;
  description: string;
  projectStatusId: number;
  status?: Status; 
  timeEstimation?: number;
  createdAt?: string;
  updatedAt?: string;
  createdBy?: number;
  roles?: Role[]; 
}

