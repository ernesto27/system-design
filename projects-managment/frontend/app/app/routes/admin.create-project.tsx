import React, { useState, useEffect } from 'react';
import { Link } from 'react-router';
import type { ProjectStatus, Role, Project } from '../types'; // Updated import
import { fetchProjectStatuses, fetchRoles, createProject } from '../api';
import ProjectRoleSelector from '../components/ProjectRoleSelector';

// Admin navigation items
const sidebarNavItems = [
  { name: 'Dashboard', href: '/admin', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6' },
  { name: 'Projects', href: '/admin/projects', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01' },
  { name: 'Create Project', href: '/admin/create-project', icon: 'M12 4v16m8-8H4', current: true },
  { name: 'Users', href: '/admin/users', icon: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z' },
  { name: 'Settings', href: '/admin/settings', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z' },
];

export default function CreateProject() {
  const [projectName, setProjectName] = useState('');
  const [description, setDescription] = useState('');
  const [timeEstimation, setTimeEstimation] = useState<number | string>('');
  const [statusId, setStatusId] = useState<number | string>(''); // Use ID for status
  const [projectStatuses, setProjectStatuses] = useState<ProjectStatus[]>([]);
  const [statusesLoading, setStatusesLoading] = useState(true);
  const [statusesError, setStatusesError] = useState<string | null>(null);
  
  // Updated state variable for role selection to use Role type directly
  const [roles, setRoles] = useState<Role[]>([]);
  const [projectRoles, setProjectRoles] = useState<Role[]>([]);
  const [rolesLoading, setRolesLoading] = useState(true);
  const [rolesError, setRolesError] = useState<string | null>(null);
  
  const [sidebarOpen, setSidebarOpen] = useState(false);
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
        description,
        projectStatusId: Number(statusId),
        timeEstimation: Number(timeEstimation),
        Roles: projectRoles.length > 0 ? projectRoles : undefined
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

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 flex flex-col">
      {/* Header */}
      <header className="bg-white dark:bg-gray-800 shadow-sm z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            {/* Mobile menu button */}
            <button
              type="button"
              className="md:hidden inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 focus:outline-none"
              onClick={() => setSidebarOpen(!sidebarOpen)}
            >
              <span className="sr-only">Open sidebar</span>
              <svg className="block h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>

            {/* Logo */}
            <div className="flex-shrink-0 flex items-center">
              <Link to="/admin" className="text-xl font-bold text-indigo-600 dark:text-indigo-500">
                PM System
              </Link>
            </div>

            {/* User menu */}
            <div className="ml-4 flex items-center md:ml-6">
              <div className="ml-3 relative">
                <div className="flex items-center">
                  <span className="hidden md:block text-sm text-gray-700 dark:text-gray-300 mr-2">Admin User</span>
                  <div className="h-8 w-8 rounded-full bg-indigo-600 flex items-center justify-center text-white">
                    A
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </header>

      <div className="flex flex-1">
        {/* Sidebar for mobile */}
        <div className={`md:hidden fixed inset-0 flex z-40 ${sidebarOpen ? 'block' : 'hidden'}`}>
          <div className="fixed inset-0 bg-gray-600 bg-opacity-75" onClick={() => setSidebarOpen(false)}></div>
          <div className="relative flex-1 flex flex-col max-w-xs w-full bg-white dark:bg-gray-800">
            <div className="absolute top-0 right-0 -mr-12 pt-2">
              <button
                type="button"
                className="ml-1 flex items-center justify-center h-10 w-10 rounded-full focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white"
                onClick={() => setSidebarOpen(false)}
              >
                <span className="sr-only">Close sidebar</span>
                <svg className="h-6 w-6 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <div className="flex-1 h-0 pt-5 pb-4 overflow-y-auto">
              <div className="flex-shrink-0 flex items-center px-4">
                <Link to="/admin" className="text-xl font-bold text-indigo-600 dark:text-indigo-500">
                  PM System
                </Link>
              </div>
              <nav className="mt-5 px-2 space-y-1">
                {sidebarNavItems.map((item) => (
                  <Link
                    key={item.name}
                    to={item.href}
                    className={`${
                      item.current ? 'bg-indigo-50 dark:bg-indigo-900 text-indigo-700 dark:text-indigo-300' : 'text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
                    } group flex items-center px-2 py-2 text-base font-medium rounded-md`}
                  >
                    <svg 
                      className={`${
                        item.current ? 'text-indigo-500 dark:text-indigo-400' : 'text-gray-400 dark:text-gray-500 group-hover:text-gray-500 dark:group-hover:text-gray-400'
                      } mr-4 flex-shrink-0 h-6 w-6`}
                      xmlns="http://www.w3.org/2000/svg" 
                      fill="none" 
                      viewBox="0 0 24 24" 
                      stroke="currentColor"
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d={item.icon} />
                    </svg>
                    {item.name}
                  </Link>
                ))}
              </nav>
            </div>
          </div>
        </div>

        {/* Static sidebar for desktop */}
        <div className="hidden md:flex md:flex-shrink-0">
          <div className="flex flex-col w-64">
            <div className="flex flex-col h-0 flex-1 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700">
              <div className="flex-1 flex flex-col pt-5 pb-4 overflow-y-auto">
                <nav className="mt-5 flex-1 px-2 space-y-1">
                  {sidebarNavItems.map((item) => (
                    <Link
                      key={item.name}
                      to={item.href}
                      className={`${
                        item.current ? 'bg-indigo-50 dark:bg-indigo-900 text-indigo-700 dark:text-indigo-300' : 'text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
                      } group flex items-center px-2 py-2 text-sm font-medium rounded-md`}
                    >
                      <svg 
                        className={`${
                          item.current ? 'text-indigo-500 dark:text-indigo-400' : 'text-gray-400 dark:text-gray-500 group-hover:text-gray-500 dark:group-hover:text-gray-400'
                        } mr-3 flex-shrink-0 h-6 w-6`}
                        xmlns="http://www.w3.org/2000/svg" 
                        fill="none" 
                        viewBox="0 0 24 24" 
                        stroke="currentColor"
                      >
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d={item.icon} />
                      </svg>
                      {item.name}
                    </Link>
                  ))}
                </nav>
              </div>
            </div>
          </div>
        </div>

        {/* Main content */}
        <div className="flex-1 flex flex-col overflow-y-auto">
          <main className="flex-1 py-6 px-4 sm:pl-4 lg:pl-6 pr-4 sm:pr-6 lg:pr-12">
            <div className="max-w-full w-full space-y-8 bg-white dark:bg-gray-800 p-12 rounded-xl shadow-lg">
              <div>
                <h2 className="mt-2 text-center text-3xl font-extrabold text-gray-900 dark:text-white">
                  Create New Project
                </h2>
                <p className="mt-3 text-center text-lg text-gray-600 dark:text-gray-400">
                  Fill in the details to create a new project in the system
                </p>
              </div>
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
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}
