import React, { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router';
import { fetchProjectById, fetchProjectStatuses, fetchRoles, updateProject } from '../api';
import type { Project, ProjectStatus, Role } from '../types';
import AdminLayout from '../components/AdminLayout';
import ProjectRoleSelector from '../components/ProjectRoleSelector';

export default function ProjectDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [project, setProject] = useState<Project | null>(null);
  const [statuses, setStatuses] = useState<ProjectStatus[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isViewMode, setIsViewMode] = useState(true);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  
  // Form state for editing
  const [projectName, setProjectName] = useState('');
  const [description, setDescription] = useState('');
  const [timeEstimation, setTimeEstimation] = useState<number | string>('');
  const [statusId, setStatusId] = useState<number | string>('');
  const [projectRoles, setProjectRoles] = useState<Role[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formErrors, setFormErrors] = useState<{ 
    projectName?: string; 
    description?: string; 
    timeEstimation?: string; 
    statusId?: string; 
    projectRoles?: string;
    api?: string 
  }>({});

  useEffect(() => {
    const loadData = async () => {
      try {
        setLoading(true);
        // Fetch project and statuses in parallel
        if (!id) {
          throw new Error('Project ID is missing');
        }
        const projectId = parseInt(id, 10);
        if (isNaN(projectId)) {
          throw new Error('Invalid project ID');
        }

        const [projectData, statusesData, rolesData] = await Promise.all([
          fetchProjectById(projectId),
          fetchProjectStatuses(),
          fetchRoles()
        ]);
        
        setProject(projectData);
        setStatuses(statusesData);
        setRoles(rolesData);
        
        // Initialize form state with project data
        setProjectName(projectData.name);
        setDescription(projectData.description);
        setTimeEstimation(projectData.timeEstimation ?? '');
        setStatusId(projectData.status?.id ?? '');
        
        // Use the roles directly from the API - they already have IDs
        setProjectRoles(projectData.roles || []);
      } catch (err) {
        console.error("Error loading project:", err);
        setError(err instanceof Error ? err.message : 'Failed to load project');
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [id]);

  // Format date function
  const formatDate = (dateString?: string) => {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const newErrors: { 
      projectName?: string; 
      description?: string; 
      timeEstimation?: string; 
      statusId?: string; 
      projectRoles?: string;
      api?: string 
    } = {};
    
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

    setFormErrors(newErrors);

    // Return if there are validation errors
    if (Object.keys(newErrors).length > 0) {
      return;
    }

    setIsSubmitting(true);
    setFormErrors({});

    try {
      if (!id) throw new Error('Project ID is missing');
      
      const projectId = parseInt(id, 10);
      if (isNaN(projectId)) throw new Error('Invalid project ID');
      
      // Create updated project object
      const updatedProject: Project = {
        id: projectId,
        name: projectName,
        description: description,
        status: { id: Number(statusId), "name": "" }, // Change from projectStatusId to status object with id property
        timeEstimation: Number(timeEstimation),
        roles: projectRoles,
        createdAt: project?.createdAt,
        updatedAt: project?.updatedAt,
        createdBy: project?.createdBy
      };
      
      const result = await updateProject(projectId, updatedProject);
      setProject(result);
      setSuccessMessage('Project updated successfully!');
      
      // Switch back to view mode after successful update
      setTimeout(() => {
        setIsViewMode(true);
        setSuccessMessage(null);
      }, 3000);
    } catch (err) {
      console.error("Error updating project:", err);
      setFormErrors({
        api: err instanceof Error ? err.message : 'Failed to update project'
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  // View mode content
  const viewContent = project && (
    <div className="bg-white dark:bg-gray-900 shadow overflow-hidden sm:rounded-lg">
      {/* Header with actions */}
      <div className="border-b border-gray-200 dark:border-gray-700 px-4 py-5 sm:px-6 flex justify-between items-center">
        <div>
          <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-white">Project Details</h3>
          <p className="mt-1 max-w-2xl text-sm text-gray-500 dark:text-gray-400">
            Complete information about the project.
          </p>
        </div>
        <div className="flex space-x-3">
          <Link to="/admin/projects" className="inline-flex items-center px-3 py-1.5 border border-gray-300 dark:border-gray-600 shadow-sm text-sm font-medium rounded-md text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 dark:focus:ring-offset-gray-900">
            <svg className="-ml-0.5 mr-2 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            Back
          </Link>
          <button
            onClick={() => setIsViewMode(false)}
            className="inline-flex items-center px-3 py-1.5 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 dark:focus:ring-offset-gray-900"
          >
            <svg className="-ml-0.5 mr-2 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
            Edit
          </button>
        </div>
      </div>
      
      {/* Project information */}
      <div className="border-t border-gray-200 dark:border-gray-700 px-4 py-5 sm:p-0">
        <dl className="sm:divide-y sm:divide-gray-200 dark:sm:divide-gray-700">
          {/* Project name */}
          <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Project name</dt>
            <dd className="mt-1 text-sm text-gray-900 dark:text-white sm:mt-0 sm:col-span-2">{project.name}</dd>
          </div>
          
          {/* Project description */}
          <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Description</dt>
            <dd className="mt-1 text-sm text-gray-900 dark:text-white sm:mt-0 sm:col-span-2">
              {project.description}
            </dd>
          </div>
          
          {/* Project status */}
          <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Status</dt>
            <dd className="mt-1 text-sm sm:mt-0 sm:col-span-2">
              {project.status && project.status.name}
            </dd>
          </div>
          
          {/* Time estimation */}
          <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Time estimation</dt>
            <dd className="mt-1 text-sm text-gray-900 dark:text-white sm:mt-0 sm:col-span-2">
              {project.timeEstimation ? `${project.timeEstimation} hours` : 'Not specified'}
            </dd>
          </div>
          
          {/* Created/Updated dates */}
          <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Created at</dt>
            <dd className="mt-1 text-sm text-gray-900 dark:text-white sm:mt-0 sm:col-span-2">
              {formatDate(project.createdAt)}
            </dd>
          </div>
          
          <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Last updated</dt>
            <dd className="mt-1 text-sm text-gray-900 dark:text-white sm:mt-0 sm:col-span-2">
              {formatDate(project.updatedAt)}
            </dd>
          </div>
          
          {/* Roles */}
          <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Roles</dt>
            <dd className="mt-1 text-sm text-gray-900 dark:text-white sm:mt-0 sm:col-span-2">
              {project.roles && project.roles.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {project.roles.map((role, index) => (
                    <span
                      key={index}
                      className="inline-flex items-center rounded-full bg-indigo-100 dark:bg-indigo-900/40 px-3 py-1 text-sm font-medium text-indigo-800 dark:text-indigo-200 border border-indigo-200 dark:border-indigo-700"
                    >
                      {role.name}
                      {role.percentage && (
                        <span className="ml-1.5 bg-indigo-200 dark:bg-indigo-700 text-indigo-800 dark:text-indigo-200 text-xs font-semibold rounded-full px-2 py-0.5">
                          {role.percentage}%
                        </span>
                      )}
                    </span>
                  ))}
                </div>
              ) : (
                <span className="text-gray-500 dark:text-gray-400">No roles assigned</span>
              )}
            </dd>
          </div>
        </dl>
      </div>
    </div>
  );

  // Edit mode content
  const editContent = project && (
    <div className="bg-white dark:bg-gray-900 shadow overflow-hidden sm:rounded-lg">
      {/* Header with actions */}
      <div className="border-b border-gray-200 dark:border-gray-700 px-4 py-5 sm:px-6 flex justify-between items-center">
        <div>
          <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-white">Edit Project</h3>
          <p className="mt-1 max-w-2xl text-sm text-gray-500 dark:text-gray-400">
            Update project information.
          </p>
        </div>
        <div className="flex space-x-3">
          <button
            onClick={() => setIsViewMode(true)}
            className="inline-flex items-center px-3 py-1.5 border border-gray-300 dark:border-gray-600 shadow-sm text-sm font-medium rounded-md text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 dark:focus:ring-offset-gray-900"
          >
            <svg className="-ml-0.5 mr-2 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
            Cancel
          </button>
        </div>
      </div>
      
      {/* Success Message */}
      {successMessage && (
        <div className="m-4 rounded-lg bg-green-50 dark:bg-green-900 p-4 shadow-md transition-all duration-300 ease-in-out">
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
      
      {/* Project edit form */}
      <form onSubmit={handleSubmit} className="border-t border-gray-200 dark:border-gray-700 px-4 py-5">
        <div className="space-y-8">
          {/* Project name */}
          <div>
            <label htmlFor="project-name" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Project Name
            </label>
            <input
              id="project-name"
              name="name"
              type="text"
              value={projectName}
              onChange={(e) => setProjectName(e.target.value)}
              className={`appearance-none rounded-md relative block w-full px-3 py-2 border ${formErrors.projectName ? "border-red-500 dark:border-red-500" : "border-gray-300 dark:border-gray-600"} placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-700 focus:outline-none focus:ring-indigo-500 dark:focus:ring-indigo-600 focus:border-indigo-500 dark:focus:border-indigo-600 focus:z-10 text-sm`}
              placeholder="Enter project name"
            />
            {formErrors.projectName && (
              <p className="mt-2 text-sm text-red-600 dark:text-red-400">
                {formErrors.projectName}
              </p>
            )}
          </div>
          
          {/* Project description */}
          <div>
            <label htmlFor="project-description" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Description
            </label>
            <textarea
              id="project-description"
              name="description"
              rows={5}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className={`appearance-none rounded-md relative block w-full px-3 py-2 border ${formErrors.description ? "border-red-500 dark:border-red-500" : "border-gray-300 dark:border-gray-600"} placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-700 focus:outline-none focus:ring-indigo-500 dark:focus:ring-indigo-600 focus:border-indigo-500 dark:focus:border-indigo-600 focus:z-10 text-sm`}
              placeholder="Describe the project and its objectives"
            />
            {formErrors.description && (
              <p className="mt-2 text-sm text-red-600 dark:text-red-400">
                {formErrors.description}
              </p>
            )}
          </div>
          
          {/* Time estimation */}
          <div>
            <label htmlFor="time-estimation" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Time Estimation (hours)
            </label>
            <input
              id="time-estimation"
              name="timeEstimation"
              type="number"
              min="1"
              value={timeEstimation}
              onChange={(e) => setTimeEstimation(e.target.value)}
              className={`appearance-none rounded-md relative block w-full px-3 py-2 border ${formErrors.timeEstimation ? "border-red-500 dark:border-red-500" : "border-gray-300 dark:border-gray-600"} placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-700 focus:outline-none focus:ring-indigo-500 dark:focus:ring-indigo-600 focus:border-indigo-500 dark:focus:border-indigo-600 focus:z-10 text-sm`}
              placeholder="Enter estimated hours to complete"
            />
            {formErrors.timeEstimation && (
              <p className="mt-2 text-sm text-red-600 dark:text-red-400">
                {formErrors.timeEstimation}
              </p>
            )}
          </div>
          
          {/* Status */}
          <div>
            <label htmlFor="project-status" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Status
            </label>
            <select
              id="project-status"
              name="status"
              value={statusId}
              onChange={(e) => setStatusId(e.target.value)}
              className={`appearance-none rounded-md relative block w-full px-3 py-2 border ${formErrors.statusId ? "border-red-500 dark:border-red-500" : "border-gray-300 dark:border-gray-600"} text-gray-900 dark:text-white bg-white dark:bg-gray-700 focus:outline-none focus:ring-indigo-500 dark:focus:ring-indigo-600 focus:border-indigo-500 dark:focus:border-indigo-600 focus:z-10 text-sm`}
            >
              <option value="">Select a status</option>
              {statuses.map((status) => (
                <option key={status.id} value={status.id} className="text-gray-900 dark:text-white bg-white dark:bg-gray-800">
                  {status.name}
                </option>
              ))}
            </select>
            {formErrors.statusId && (
              <p className="mt-2 text-sm text-red-600 dark:text-red-400">
                {formErrors.statusId}
              </p>
            )}
          </div>
          
          {/* Project Roles */}
          <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
            <ProjectRoleSelector
              roles={roles}
              projectRoles={projectRoles}
              setProjectRoles={setProjectRoles}
              loading={false}
              error={null}
            />
          </div>
          
          {/* API Error */}
          {formErrors.api && (
            <div className="rounded-md bg-red-50 dark:bg-red-900/30 p-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-red-800 dark:text-red-200">Error</h3>
                  <div className="mt-2 text-sm text-red-700 dark:text-red-300">
                    <p>{formErrors.api}</p>
                  </div>
                </div>
              </div>
            </div>
          )}
          
          {/* Submit button */}
          <div className="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              onClick={() => setIsViewMode(true)}
              className="inline-flex justify-center py-2 px-4 border border-gray-300 dark:border-gray-600 shadow-sm text-sm font-medium rounded-md text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 dark:focus:ring-offset-gray-900"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 dark:focus:ring-offset-gray-900 disabled:opacity-50"
            >
              {isSubmitting ? "Saving..." : "Save Changes"}
            </button>
          </div>
        </div>
      </form>
    </div>
  );

  // Content to render inside the layout
  const content = loading ? (
    <div className="text-center py-8">
      <svg className="animate-spin h-10 w-10 text-indigo-500 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
      <p className="mt-4 text-gray-500 dark:text-gray-400">Loading project details...</p>
    </div>
  ) : error ? (
    <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
      <p>Error: {error}</p>
      <div className="mt-4">
        <Link to="/admin/projects" className="text-indigo-600 hover:text-indigo-800 dark:text-indigo-400 dark:hover:text-indigo-300">
          &larr; Back to Projects
        </Link>
      </div>
    </div>
  ) : !project ? (
    <div className="text-center py-8">
      <svg className="h-16 w-16 text-gray-400 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
      <p className="mt-4 text-gray-500 dark:text-gray-400">Project not found.</p>
      <Link to="/admin/projects" className="mt-4 inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
        Back to Projects
      </Link>
    </div>
  ) : isViewMode ? viewContent : editContent;

  return (
    <AdminLayout
      title={project ? project.name : 'Project Details'}
      subtitle={project ? (isViewMode ? `View details for ${project.name}` : `Edit ${project.name}`) : 'Loading project information'}
      currentPath="/projects"
    >
      {content}
    </AdminLayout>
  );
}