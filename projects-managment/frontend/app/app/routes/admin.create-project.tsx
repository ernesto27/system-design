import React, { useState, useEffect } from 'react';
import { Link } from 'react-router';
import type { ProjectStatus, Role, Project } from '../types';
import { fetchProjectStatuses, fetchRoles, createProject } from '../api';
import ProjectRoleSelector from '../components/ProjectRoleSelector';
import AdminLayout from '../components/AdminLayout';

export default function CreateProject() {
  const [projectName, setProjectName] = useState('');
  const [description, setDescription] = useState('');
  const [timeEstimation, setTimeEstimation] = useState<number | string>('');
  const [statusId, setStatusId] = useState<number | string>('');
  const [projectStatuses, setProjectStatuses] = useState<ProjectStatus[]>([]);
  const [statusesLoading, setStatusesLoading] = useState(true);
  const [statusesError, setStatusesError] = useState<string | null>(null);
  
  const [roles, setRoles] = useState<Role[]>([]);
  const [projectRoles, setProjectRoles] = useState<Role[]>([]);
  const [rolesLoading, setRolesLoading] = useState(true);
  const [rolesError, setRolesError] = useState<string | null>(null);
  
  const [errors, setErrors] = useState<{ projectName?: string; description?: string; timeEstimation?: string; statusId?: string; projectRoles?: string; api?: string }>({});
  const [isLoading, setIsLoading] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  // Fetch project statuses and roles on component mount
  useEffect(() => {
    const loadData = async () => {
      try {
        // Fetch project statuses
        setStatusesLoading(true);
        const statusesData = await fetchProjectStatuses();
        setProjectStatuses(statusesData);
        if (statusesData.length > 0) {
          setStatusId(statusesData[0].id); // Set default status
        }
        setStatusesError(null);
      } catch (error) {
        console.error("Failed to fetch statuses:", error);
        setStatusesError(error instanceof Error ? error.message : "Could not load statuses.");
      } finally {
        setStatusesLoading(false);
      }

      try {
        // Fetch roles
        setRolesLoading(true);
        const rolesData = await fetchRoles();
        setRoles(rolesData);
        setRolesError(null);
      } catch (error) {
        console.error("Failed to fetch roles:", error);
        setRolesError(error instanceof Error ? error.message : "Could not load roles.");
      } finally {
        setRolesLoading(false);
      }
    };

    loadData();
  }, []);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const newErrors: { projectName?: string; description?: string; timeEstimation?: string; statusId?: string; projectRoles?: string; api?: string } = {};
    
    // Clear previous success message
    setSuccessMessage(null);

    // Validation logic
    if (!projectName.trim()) {
      newErrors.projectName = "Project name is required.";
    } else if (projectName.trim().length < 3) {
      newErrors.projectName = "Project name must be at least 3 characters.";
    }

    if (!description.trim()) {
      newErrors.description = "Project description is required.";
    } else if (description.trim().length < 10) {
      newErrors.description = "Description must be at least 10 characters.";
    }

    if (!timeEstimation) {
      newErrors.timeEstimation = "Time estimation is required.";
    } else if (Number(timeEstimation) <= 0) {
      newErrors.timeEstimation = "Time estimation must be greater than 0.";
    }

    if (!statusId) {
      newErrors.statusId = "Project status is required.";
    }

    setErrors(newErrors);

    // Scroll to the first input with an error if there are validation errors
    if (Object.keys(newErrors).length > 0) {
      const errorFields = ['projectName', 'description', 'timeEstimation', 'statusId', 'projectRoles'];
      for (const field of errorFields) {
        if (newErrors[field as keyof typeof newErrors]) {
          const element = document.getElementById(field === 'projectName' ? 'project-name' : 
                                                 field === 'description' ? 'project-description' : 
                                                 field === 'timeEstimation' ? 'time-estimation' : 
                                                 field === 'statusId' ? 'project-status' : 
                                                 field === 'projectRoles' ? 'role-select' : '');
          if (element) {
            element.scrollIntoView({ behavior: 'smooth', block: 'center' });
            element.focus();
            break; 
          }
        }
      }
      return; 
    }

    setIsLoading(true);
    setErrors({}); // Clear previous API errors

    try {
      // Create a project object that follows the Project interface structure
      const projectData: Project = {
        name: projectName,
        description: description,
        projectStatusId: Number(statusId),
        timeEstimation: Number(timeEstimation),
        roles: projectRoles.length > 0 ? projectRoles : undefined
      };
      
      const createdProject = await createProject(projectData);

      console.log('Project created:', createdProject);
      
      // Set success message instead of alert
      setSuccessMessage(`Project was created successfully!`);

      // Reset form
      setProjectName('');
      setDescription('');
      setTimeEstimation('');
      setStatusId(projectStatuses.length > 0 ? projectStatuses[0].id : '');
      setProjectRoles([]); // Reset project roles

      // Auto-hide success message after 5 seconds
      setTimeout(() => {
        setSuccessMessage(null);
      }, 5000);

    } catch (error) {
      console.error("Project creation failed:", error);
      setErrors({ 
        api: error instanceof Error ? error.message : "Failed to create project. Please try again later." 
      });
      
      // Scroll to the API error message if there is one
      const apiErrorElement = document.getElementById('api-error');
      if (apiErrorElement) {
        apiErrorElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
      }
    } finally {
      setIsLoading(false);
    }
  };

  // Content to render inside the layout
  const formContent = (
    <form className="mt-12 space-y-10" onSubmit={handleSubmit} noValidate>
      <div className="rounded-md shadow-sm space-y-8">
        <div>
          <label htmlFor="project-name" className="block text-base font-medium text-gray-700 dark:text-gray-300 mb-3">
            Project Name
          </label>
          <input
            id="project-name"
            name="name"
            type="text"
            value={projectName}
            onChange={(e) => setProjectName(e.target.value)}
            className={`appearance-none rounded-lg relative block w-full px-4 py-3 border ${errors.projectName ? "border-red-500 dark:border-red-500" : "border-gray-300 dark:border-gray-600"} placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-700 focus:outline-none focus:ring-indigo-500 dark:focus:ring-indigo-600 focus:border-indigo-500 dark:focus:border-indigo-600 focus:z-10 text-base`}
            placeholder="Enter project name"
            aria-invalid={errors.projectName ? "true" : "false"}
            aria-describedby={errors.projectName ? "projectName-error" : undefined}
          />
          {errors.projectName && (
            <p id="projectName-error" className="mt-2 text-sm text-red-600 dark:text-red-400">
              {errors.projectName}
            </p>
          )}
        </div>
        <div>
          <label htmlFor="project-description" className="block text-base font-medium text-gray-700 dark:text-gray-300 mb-3">
            Description
          </label>
          <textarea
            id="project-description"
            name="description"
            rows={8}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className={`appearance-none rounded-lg relative block w-full px-4 py-3 border ${errors.description ? "border-red-500 dark:border-red-500" : "border-gray-300 dark:border-gray-600"} placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-700 focus:outline-none focus:ring-indigo-500 dark:focus:ring-indigo-600 focus:border-indigo-500 dark:focus:border-indigo-600 focus:z-10 text-base`}
            placeholder="Describe the project and its objectives"
            aria-invalid={errors.description ? "true" : "false"}
            aria-describedby={errors.description ? "description-error" : undefined}
          />
          {errors.description && (
            <p id="description-error" className="mt-2 text-sm text-red-600 dark:text-red-400">
              {errors.description}
            </p>
          )}
        </div>
        <div>
          <label htmlFor="time-estimation" className="block text-base font-medium text-gray-700 dark:text-gray-300 mb-3">
            Time Estimation (hours)
          </label>
          <input
            id="time-estimation"
            name="timeEstimation"
            type="number"
            min="1"
            value={timeEstimation}
            onChange={(e) => setTimeEstimation(e.target.value)}
            className={`appearance-none rounded-lg relative block w-full px-4 py-3 border ${errors.timeEstimation ? "border-red-500 dark:border-red-500" : "border-gray-300 dark:border-gray-600"} placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-700 focus:outline-none focus:ring-indigo-500 dark:focus:ring-indigo-600 focus:border-indigo-500 dark:focus:border-indigo-600 focus:z-10 text-base`}
            placeholder="Enter estimated hours to complete"
            aria-invalid={errors.timeEstimation ? "true" : "false"}
            aria-describedby={errors.timeEstimation ? "timeEstimation-error" : undefined}
          />
          {errors.timeEstimation && (
            <p id="timeEstimation-error" className="mt-2 text-sm text-red-600 dark:text-red-400">
              {errors.timeEstimation}
            </p>
          )}
        </div>
        <div>
          <label htmlFor="project-status" className="block text-base font-medium text-gray-700 dark:text-gray-300 mb-3">
            Status
          </label>
          <select
            id="project-status"
            name="status"
            value={statusId}
            onChange={(e) => setStatusId(e.target.value)}
            disabled={statusesLoading || !!statusesError}
            className={`appearance-none rounded-lg relative block w-full px-4 py-3 border ${errors.statusId ? "border-red-500 dark:border-red-500" : "border-gray-300 dark:border-gray-600"} text-gray-900 dark:text-white bg-white dark:bg-gray-700 focus:outline-none focus:ring-indigo-500 dark:focus:ring-indigo-600 focus:border-indigo-500 dark:focus:border-indigo-600 focus:z-10 text-base disabled:opacity-50 disabled:cursor-not-allowed`}
            aria-invalid={errors.statusId ? "true" : "false"}
            aria-describedby={errors.statusId ? "statusId-error" : undefined}
          >
            {statusesLoading && <option value="">Loading statuses...</option>}
            {statusesError && <option value="">Error loading statuses</option>}
            {!statusesLoading && !statusesError && projectStatuses.length === 0 && <option value="">No statuses available</option>}
            {!statusesLoading && !statusesError && projectStatuses.map((status) => (
              <option key={status.id} value={status.id} className="text-gray-900 dark:text-white bg-white dark:bg-gray-800">
                {status.name}
              </option>
            ))}
          </select>
          {errors.statusId && (
            <p id="statusId-error" className="mt-2 text-sm text-red-600 dark:text-red-400">
              {errors.statusId}
            </p>
          )}
          {statusesError && (
            <p className="mt-2 text-sm text-red-600 dark:text-red-400">
              {statusesError}
            </p>
          )}
        </div>
        
        {/* Project Roles Section */}
        <div className="border-t border-gray-200 dark:border-gray-700 pt-8">
          <ProjectRoleSelector
            roles={roles}
            projectRoles={projectRoles}
            setProjectRoles={setProjectRoles}
            loading={rolesLoading}
            error={rolesError}
          />
          {errors.projectRoles && (
            <p className="mt-2 text-sm text-red-600 dark:text-red-400">
              {errors.projectRoles}
            </p>
          )}
        </div>
      </div>

      {errors.api && (
        <p id="api-error" className="text-sm text-red-600 dark:text-red-400 text-center">
          {errors.api}
        </p>
      )}

      {/* Success Message */}
      {successMessage && (
        <div className="mb-4 rounded-lg bg-green-50 dark:bg-green-900 p-4 shadow-md transition-all duration-300 ease-in-out">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-green-600 dark:text-green-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <p className="text-sm font-medium text-green-800 dark:text-green-200">
                {successMessage}
              </p>
            </div>
            <div className="ml-auto pl-3">
              <div className="-mx-1.5 -my-1.5">
                <button
                  type="button"
                  onClick={() => setSuccessMessage(null)}
                  className="inline-flex rounded-md p-1.5 text-green-700 dark:text-green-300 hover:bg-green-100 dark:hover:bg-green-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 dark:focus:ring-offset-green-800"
                >
                  <span className="sr-only">Dismiss</span>
                  <svg className="h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                    <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      <div className="pt-6">
        <button
          type="submit"
          disabled={isLoading || statusesLoading}
          className="group relative w-full flex justify-center py-4 px-6 border border-transparent text-lg font-medium rounded-lg text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 dark:focus:ring-offset-gray-800 transition-colors shadow-md disabled:opacity-50"
        >
          {isLoading ? "Creating Project..." : "Create Project"}
        </button>
      </div>
    </form>
  );

  return (
    <AdminLayout 
      title="Create New Project" 
      subtitle="Fill in the details to create a new project in the system"
      currentPath="/admin/create-project"
    >
      {formContent}
    </AdminLayout>
  );
}
