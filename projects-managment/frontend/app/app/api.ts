import type { ProjectStatus, Project } from './types';

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
 * Creates a new project.
 */
    export const createProject = async (projectData: { 
  name: string; 
  description: string; 
  project_status_id: number;
  time_estimation?: number;
}): Promise<Project> => {
  const response = await fetchWithAuth(`${API_BASE_URL}/projects`, {
    method: 'POST',
    body: JSON.stringify(projectData),
  });
  return response.json();
};

// Add other API functions as needed (e.g., fetchProjects, loginUser, etc.)
