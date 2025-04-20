import type { ProjectStatus, Project, Role } from './types';

const API_BASE_URL = 'http://localhost:8080/api/v1';

// Helper function to get the auth token
const getAuthToken = (): string | null => {
  return localStorage.getItem('authToken') || sessionStorage.getItem('authToken');
};

// Helper function for making authenticated requests
const fetchWithAuth = async (url: string, options: RequestInit = {}): Promise<Response> => {
  const token = getAuthToken();
  if (!token) {
    throw new Error("No authentication token found");
  }

  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
    ...(options.headers || {}),
  };

  const response = await fetch(url, {
    ...options,
    headers,
  });

  if (!response.ok) {
    // Try to parse error message from backend
    let errorMessage = `API request failed with status ${response.status}`;
    try {
      const errorData = await response.json();
      errorMessage = errorData.message || errorData.error || errorMessage;
    } catch (e) {
      // Ignore if response is not JSON or empty
    }
    throw new Error(errorMessage);
  }

  return response;
};


/**
 * Fetches all available project statuses.
 */
export const fetchProjectStatuses = async (): Promise<ProjectStatus[]> => {
  const response = await fetch(`${API_BASE_URL}/project-statuses`); 
  if (!response.ok) {
    let errorMessage = 'Failed to fetch project statuses';
    try {
      const errorData = await response.json();
      errorMessage = errorData.message || errorData.error || errorMessage;
    } catch (e) { /* Ignore if response body is not JSON */ }
    throw new Error(errorMessage);
  }
  
  const responseData = await response.json(); 

  if (responseData && Array.isArray(responseData.data)) {
    return responseData.data as ProjectStatus[];
  } else if (Array.isArray(responseData)) {
    return responseData as ProjectStatus[];
  } else {
    console.error("Unexpected API response structure for project statuses:", responseData);
    throw new Error("Received unexpected data format for project statuses.");
  }
};

/**
 * Fetches all available roles.
 */
export const fetchRoles = async (): Promise<Role[]> => {
  const response = await fetchWithAuth(`${API_BASE_URL}/roles`);
  
  const responseData = await response.json(); 

  if (responseData && Array.isArray(responseData.data)) {
    return responseData.data as Role[];
  } else if (Array.isArray(responseData)) {
    return responseData as Role[];
  } else {
    console.error("Unexpected API response structure for roles:", responseData);
    throw new Error("Received unexpected data format for roles.");
  }
};

/**
 * Creates a new project.
 */
export const createProject = async (project: Project): Promise<Project> => {
  const response = await fetchWithAuth(`${API_BASE_URL}/projects`, {
    method: 'POST',
    body: JSON.stringify(project),
  });
  return response.json();
};

/**
 * Fetches all projects
 */
export const fetchProjects = async (): Promise<Project[]> => {
  const response = await fetchWithAuth(`${API_BASE_URL}/projects`);
  
  const responseData = await response.json(); 

  if (responseData && Array.isArray(responseData.data)) {
    return responseData.data as Project[];
  } else if (Array.isArray(responseData)) {
    return responseData as Project[];
  } else {
    console.error("Unexpected API response structure for projects:", responseData);
    throw new Error("Received unexpected data format for projects.");
  }
};

/**
 * Fetches a project by its ID
 */
export const fetchProjectById = async (id: number): Promise<Project> => {
  const response = await fetchWithAuth(`${API_BASE_URL}/projects/${id}`);
  
  const responseData = await response.json(); 

  if (responseData && responseData.data) {
    return responseData.data as Project;
  } else if (responseData && !responseData.data) {
    return responseData as Project;
  } else {
    console.error("Unexpected API response structure for project details:", responseData);
    throw new Error("Received unexpected data format for project details.");
  }
};

/**
 * Updates an existing project.
 */
export const updateProject = async (id: number, project: Project): Promise<Project> => {
  const response = await fetchWithAuth(`${API_BASE_URL}/projects/${id}`, {
    method: 'PUT',
    body: JSON.stringify(project),
  });
  
  const responseData = await response.json();
  
  if (responseData && responseData.data) {
    return responseData.data as Project;
  } else {
    return responseData as Project;
  }
};

// Add other API functions as needed (e.g., loginUser, etc.)
