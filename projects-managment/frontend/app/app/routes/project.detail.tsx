import React, { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router';
import { fetchProjectById, fetchProjectStatuses, fetchRoles, updateProject, fetchProjectComments, createComment, deleteComment, likeComment, unlikeComment } from '../api';
import type { Project, ProjectStatus, Role, Comment } from '../types';
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
  
  // Comments state
  const [comments, setComments] = useState<Comment[]>([]);
  const [newComment, setNewComment] = useState("");
  const [loadingComments, setLoadingComments] = useState(false);
  const [commentError, setCommentError] = useState<string | null>(null);
  const [addingComment, setAddingComment] = useState(false);
  const [processingLike, setProcessingLike] = useState<number | null>(null);
  
  // Comment deletion confirmation modal
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [commentToDelete, setCommentToDelete] = useState<number | null>(null);
  
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

        // Load comments
        loadComments(projectId);
      } catch (err) {
        console.error("Error loading project:", err);
        setError(err instanceof Error ? err.message : 'Failed to load project');
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [id]);

  // Load comments for the project
  const loadComments = async (projectId: number) => {
    try {
      setLoadingComments(true);
      setCommentError(null);
      const commentsData = await fetchProjectComments(projectId);
      setComments(commentsData as unknown as Comment[]);
    } catch (err) {
      console.error("Error loading comments:", err);
      setCommentError(err instanceof Error ? err.message : 'Failed to load comments');
    } finally {
      setLoadingComments(false);
    }
  };

  // Handle adding a new comment
  const handleAddComment = async () => {
    if (!newComment.trim() || !project?.id) return;
    
    try {
      setAddingComment(true);
      setCommentError(null);
      
      const comment: Comment = {
        projectId: project.id,
        userId: 0,
        content: newComment.trim()
      };
      
      await createComment(comment);
      setNewComment(''); // Clear input field
      
      // Reload comments to get the newly added one
      if (project.id) {
        await loadComments(project.id);
      }
    } catch (err) {
      console.error("Error adding comment:", err);
      setCommentError(err instanceof Error ? err.message : 'Failed to add comment');
    } finally {
      setAddingComment(false);
    }
  };
  
  // Show delete confirmation modal
  const showDeleteConfirmation = (commentId: number) => {
    setCommentToDelete(commentId);
    setIsDeleteModalOpen(true);
  };

  // Confirm delete and actually delete the comment
  const confirmDeleteComment = async () => {
    if (!commentToDelete) return;
    
    try {
      await deleteComment(commentToDelete);
      
      // Update local state to remove the deleted comment
      setComments(comments.filter(comment => comment.id !== commentToDelete));
      
      // Close the modal
      setIsDeleteModalOpen(false);
      setCommentToDelete(null);
    } catch (err) {
      console.error("Error deleting comment:", err);
      setCommentError(err instanceof Error ? err.message : 'Failed to delete comment');
    }
  };
  
  // Handle deleting a comment
  const handleDeleteComment = async (commentId: number) => {
    if (!commentId) return;
    
    try {
      await deleteComment(commentId);
      
      // Update local state to remove the deleted comment
      setComments(comments.filter(comment => comment.id !== commentId));
    } catch (err) {
      console.error("Error deleting comment:", err);
      setCommentError(err instanceof Error ? err.message : 'Failed to delete comment');
    }
  };

  // Handle liking a comment
  const handleLikeComment = async (commentId: number) => {
    if (!commentId) return;
    
    try {
      setProcessingLike(commentId);
      setCommentError(null);
      
      await likeComment(commentId);
      
      // Update local state to reflect the like
      setComments(comments.map(comment => {
        if (comment.id === commentId) {
          return {
            ...comment,
            likesCount: (comment.likesCount || 0) + 1,
            isLiked: true
          };
        }
        return comment;
      }));
    } catch (err) {
      console.error("Error liking comment:", err);
      setCommentError(err instanceof Error ? err.message : 'Failed to like comment');
    } finally {
      setProcessingLike(null);
    }
  };
  
  // Handle unliking a comment
  const handleUnlikeComment = async (commentId: number) => {
    if (!commentId) return;
    
    try {
      setProcessingLike(commentId);
      setCommentError(null);
      
      await unlikeComment(commentId);
      
      // Update local state to reflect the unlike
      setComments(comments.map(comment => {
        if (comment.id === commentId) {
          return {
            ...comment,
            likesCount: Math.max((comment.likesCount || 0) - 1, 0),
            isLiked: false
          };
        }
        return comment;
      }));
    } catch (err) {
      console.error("Error unliking comment:", err);
      setCommentError(err instanceof Error ? err.message : 'Failed to unlike comment');
    } finally {
      setProcessingLike(null);
    }
  };

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
      
      {/* Comments section */}
      <div className="border-t border-gray-200 dark:border-gray-700 px-6 py-5">
        <div className="mb-6">
          <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-white mb-4">Comments</h3>
          
          {/* Add comment form */}
          <div className="mb-6">
            <div className="flex">
              <textarea
                placeholder="Add a comment..."
                value={newComment}
                onChange={(e) => setNewComment(e.target.value)}
                className="flex-grow rounded-md border border-gray-300 dark:border-gray-600 shadow-sm px-4 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 dark:focus:ring-indigo-500 focus:border-indigo-500 dark:focus:border-indigo-500"
                rows={3}
              />
            </div>
            <div className="mt-2 flex justify-end">
              <button
                type="button"
                onClick={handleAddComment}
                disabled={addingComment || !newComment.trim()}
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
              >
                {addingComment ? (
                  <>
                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Posting...
                  </>
                ) : (
                  <>
                    <svg className="-ml-1 mr-2 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                    </svg>
                    Post Comment
                  </>
                )}
              </button>
            </div>
          </div>
          
          {/* Error message */}
          {commentError && (
            <div className="rounded-md bg-red-50 dark:bg-red-900/30 p-4 mb-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm font-medium text-red-800 dark:text-red-200">{commentError}</p>
                </div>
              </div>
            </div>
          )}
          
          {/* Comments list */}
          {loadingComments ? (
            <div className="text-center py-8">
              <svg className="animate-spin h-8 w-8 text-indigo-500 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">Loading comments...</p>
            </div>
          ) : comments.length === 0 ? (
            <div className="text-center py-8 border rounded-md border-gray-200 dark:border-gray-700">
              <svg className="h-12 w-12 text-gray-400 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
              </svg>
              <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">No comments yet. Be the first to comment!</p>
            </div>
          ) : (
            <div className="space-y-4">
              {comments.map((comment) => (
                <div key={comment.id} className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 shadow-sm">
                  <div className="flex justify-between items-start">
                    <div className="flex items-center">
                      <div className="w-8 h-8 rounded-full bg-indigo-100 dark:bg-indigo-800 flex items-center justify-center text-indigo-800 dark:text-indigo-100 font-semibold text-sm mr-3">
                        {comment.user?.username ? comment.user.username.charAt(0).toUpperCase() : '?'}
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-900 dark:text-white">{comment.user?.username || 'Unknown User'}</p>
                        <p className="text-xs text-gray-500 dark:text-gray-400">{formatDate(comment.createdAt)}</p>
                      </div>
                    </div>
                    <button
                      onClick={() => showDeleteConfirmation(comment.id!)}
                      className="text-gray-400 hover:text-red-500 dark:hover:text-red-400"
                      title="Delete comment"
                    >
                      <svg className="h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </div>
                  <div className="mt-3">
                    <p className="text-sm text-gray-800 dark:text-gray-200 whitespace-pre-wrap">{comment.content}</p>
                    
                    {/* Like/Unlike section */}
                    <div className="mt-3 flex items-center">
                      {comment.isLiked ? (
                        <button 
                          onClick={() => handleUnlikeComment(comment.id!)}
                          disabled={processingLike === comment.id}
                          className="inline-flex items-center text-sm text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300 disabled:opacity-50"
                        >
                          <svg className="h-5 w-5 mr-1 fill-current" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20">
                            <path fillRule="evenodd" d="M3.172 5.172a4 4 0 015.656 0L10 6.343l1.172-1.171a4 4 0 115.656 5.656L10 17.657l-6.828-6.829a4 4 0 010-5.656z" clipRule="evenodd" />
                          </svg>
                          {processingLike === comment.id ? 'Processing...' : 'Unlike'}
                        </button>
                      ) : (
                        <button 
                          onClick={() => handleLikeComment(comment.id!)}
                          disabled={processingLike === comment.id}
                          className="inline-flex items-center text-sm text-gray-500 dark:text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 disabled:opacity-50"
                        >
                          <svg className="h-5 w-5 mr-1 stroke-current" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                          </svg>
                          {processingLike === comment.id ? 'Processing...' : 'Like'}
                        </button>
                      )}
                      <span className="ml-2 text-xs text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-700 rounded-full px-2 py-1">
                        {comment.likesCount || 0} {comment.likesCount === 1 ? 'like' : 'likes'}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
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
      
      {/* Delete Comment Confirmation Modal */}
      {isDeleteModalOpen && (
        <div className="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
          <div className="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            {/* Background overlay */}
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" aria-hidden="true" onClick={() => setIsDeleteModalOpen(false)}></div>
            
            {/* Center modal properly */}
            <span className="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
            
            {/* Modal panel */}
            <div className="relative inline-block align-bottom bg-white dark:bg-gray-800 rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
              <div className="bg-white dark:bg-gray-800 px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                <div className="sm:flex sm:items-start">
                  <div className="mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full bg-red-100 dark:bg-red-900 sm:mx-0 sm:h-10 sm:w-10">
                    <svg className="h-6 w-6 text-red-600 dark:text-red-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                    </svg>
                  </div>
                  <div className="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
                    <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-white" id="modal-title">
                      Delete Comment
                    </h3>
                    <div className="mt-2">
                      <p className="text-sm text-gray-500 dark:text-gray-400">
                        Are you sure you want to delete this comment? This action cannot be undone.
                      </p>
                    </div>
                  </div>
                </div>
              </div>
              <div className="bg-gray-50 dark:bg-gray-700 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button 
                  type="button" 
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-red-600 text-base font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={confirmDeleteComment}
                >
                  Delete
                </button>
                <button 
                  type="button" 
                  className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 dark:border-gray-600 shadow-sm px-4 py-2 bg-white dark:bg-gray-800 text-base font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={() => setIsDeleteModalOpen(false)}
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </AdminLayout>
  );
}