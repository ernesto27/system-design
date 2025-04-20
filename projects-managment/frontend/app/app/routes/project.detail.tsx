import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router';
import { fetchProjectById, fetchProjectStatuses } from '../api';
import type { Project, ProjectStatus } from '../types';
import AdminLayout from '../components/AdminLayout';

export default function ProjectDetail() {
  const { id } = useParams<{ id: string }>();
  const [project, setProject] = useState<Project | null>(null);
  const [statuses, setStatuses] = useState<ProjectStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isViewMode, setIsViewMode] = useState(true); // Track whether we're in view or edit mode

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

        const [projectData, statusesData] = await Promise.all([
          fetchProjectById(projectId),
          fetchProjectStatuses()
        ]);
        
        setProject(projectData);
        setStatuses(statusesData);
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
  ) : (
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

  return (
    <AdminLayout
      title={project ? project.name : 'Project Details'}
      subtitle={project ? `View details for ${project.name}` : 'Loading project information'}
      currentPath="/projects"
    >
      {content}
    </AdminLayout>
  );
}