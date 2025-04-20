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
  status?: Status; 
  timeEstimation?: number;
  createdAt?: string;
  updatedAt?: string;
  createdBy?: number;
  roles?: Role[]; 
}

export interface User {
  id: number;
  username: string;
  password?: string;
  email?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface Comment {
  id?: number;
  projectId: number;
  userId: number;
  content: string;
  user?: User;
  likesCount?: number;
  isLiked?: boolean;
  createdAt?: string;
  updatedAt?: string;
}

