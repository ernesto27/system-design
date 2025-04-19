import React, { useState, useEffect } from 'react';
import { fetchProjects, fetchProjectStatuses } from '../api';
import type { Project, ProjectStatus } from '../types';
import AdminLayout from '../components/AdminLayout';
import { Link } from 'react-router';

export default function ProjectsList() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [statuses, setStatuses] = useState<ProjectStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadData = async () => {
      try {
        setLoading(true);
        // Fetch projects and statuses in parallel
        const [projectsData, statusesData] = await Promise.all([
          fetchProjects(),
          fetchProjectStatuses()
        ]);
        
        setProjects(projectsData);
        setStatuses(statusesData);
      } catch (err) {
        console.error("Error loading projects:", err);
        setError(err instanceof Error ? err.message : 'Failed to load projects');
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, []);

  // Helper function to get status name by ID
  const getStatusName = (statusId: number): string => {
    const status = statuses.find(s => s.id === statusId);
    return status ? status.name : 'Unknown Status';
  };

  // Content to render inside the layout
  const content = loading ? (
    <div className="text-center py-8">
      <svg className="animate-spin h-10 w-10 text-indigo-500 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
      <p className="mt-4 text-gray-500 dark:text-gray-400">Loading projects...</p>
    </div>
  ) : error ? (
    <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
      <p>Error: {error}</p>
    </div>
  ) : projects.length === 0 ? (
    <div className="text-center py-8">
      <svg className="h-16 w-16 text-gray-400 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
      <p className="mt-4 text-gray-500 dark:text-gray-400">No projects found.</p>
      <Link to="/admin/create-project" className="mt-4 inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
        Create a project
      </Link>
    </div>
  ) : (
    <div className="mt-8 flex flex-col">
      <div className="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
        <div className="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
          <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
            <table className="min-w-full divide-y divide-gray-300 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-gray-800">
                <tr>
                  <th scope="col" className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 dark:text-gray-200 sm:pl-6">Name</th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-gray-200">Description</th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-gray-200">Status</th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-gray-200">Time Est.</th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-gray-200">Roles</th>
                  <th scope="col" className="relative py-3.5 pl-3 pr-4 sm:pr-6">
                    <span className="sr-only">Actions</span>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700 bg-white dark:bg-gray-900">
                {projects.map((project, index) => (
                  <tr key={index} className="hover:bg-gray-50 dark:hover:bg-gray-800 transition duration-150">
                    <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 dark:text-white sm:pl-6">
                      <div className="flex items-center">
                        <div className="h-8 w-8 flex-shrink-0 rounded-full bg-indigo-100 dark:bg-indigo-900 flex items-center justify-center">
                          <span className="text-indigo-700 dark:text-indigo-300 font-medium">{project.name.charAt(0).toUpperCase()}</span>
                        </div>
                        <div className="ml-4">
                          <div className="font-medium text-gray-900 dark:text-white">{project.name}</div>
                        </div>
                      </div>
                    </td>
                    <td className="px-3 py-4 text-sm text-gray-500 dark:text-gray-400 max-w-md">
                      <div className="line-clamp-2">{project.description}</div>
                    </td>
                    <td className="px-3 py-4 text-sm">
                      {(() => {
                      const statusName = getStatusName(project.projectStatusId).toLowerCase();
                      // Define color classes for each status with proper type definition
                      const statusClasses: Record<string, { wrapper: string; dot: string }> = {
                        'active': {
                          wrapper: 'bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-300 border-green-200 dark:border-green-800/30',
                          dot: 'bg-green-500 dark:bg-green-400'
                        },
                        'pending': {
                          wrapper: 'bg-yellow-50 dark:bg-yellow-900/20 text-yellow-700 dark:text-yellow-300 border-yellow-200 dark:border-yellow-800/30',
                          dot: 'bg-yellow-500 dark:bg-yellow-400'
                        },
                        'completed': {
                          wrapper: 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 border-blue-200 dark:border-blue-800/30',
                          dot: 'bg-blue-500 dark:bg-blue-400'
                        },
                        'cancelled': {
                          wrapper: 'bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 border-red-200 dark:border-red-800/30',
                          dot: 'bg-red-500 dark:bg-red-400'
                        },
                        'default': {
                          wrapper: 'bg-gray-50 dark:bg-gray-800/40 text-gray-700 dark:text-gray-300 border-gray-200 dark:border-gray-700',
                          dot: 'bg-gray-500 dark:bg-gray-400'
                        }
                      };
                      // Get the classes for this status or use default if not found
                      const classes = statusClasses[statusName] || statusClasses.default;
                      return (
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium border ${classes.wrapper}`}>
                          <span className={`h-1 w-1 rounded-full mr-1 ${classes.dot}`}></span>
                          {getStatusName(project.projectStatusId)}
                        </span>
                      );
                      })()}
                    </td>
                    <td className="px-3 py-4 text-sm text-gray-500 dark:text-gray-400">
                      {project.timeEstimation ? (
                        <div className="flex items-center">
                          <svg className="mr-1.5 h-4 w-4 text-gray-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                          </svg>
                          {project.timeEstimation}h
                        </div>
                      ) : '-'}
                    </td>
                    <td className="px-3 py-4 text-sm text-gray-500 dark:text-gray-400">
                      {project.roles && project.roles.length > 0 ? (
                        <div className="flex flex-wrap gap-1.5">
                          {project.roles.map((role, roleIndex) => (
                            <span 
                              key={roleIndex}
                              className="inline-flex items-center rounded-full bg-gray-100 dark:bg-gray-700 px-2.5 py-0.5 text-xs font-medium text-gray-800 dark:text-gray-300 border border-gray-200 dark:border-gray-600"
                              title={`${role.name}: ${role.percentage}%`}
                            >
                              {role.name}
                              {role.percentage && (
                                <span className="ml-1 text-gray-500 dark:text-gray-400">{role.percentage}%</span>
                              )}
                            </span>
                          ))}
                        </div>
                      ) : '-'}
                    </td>
                    <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                      <div className="flex items-center justify-end space-x-3">
                        <button className="flex items-center text-indigo-600 dark:text-indigo-400 hover:text-indigo-900 dark:hover:text-indigo-300 transition-colors">
                          <svg className="mr-1 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                          </svg>
                          View<span className="sr-only">, {project.name}</span>
                        </button>
                        <button className="flex items-center text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 transition-colors">
                          <svg className="mr-1 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                          </svg>
                          Edit<span className="sr-only">, {project.name}</span>
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );

  return (
    <AdminLayout 
      title="Projects List" 
      subtitle="View all projects in the system"
      currentPath="/projects"
    >
      {content}
    </AdminLayout>
  );
}