import React, { useState } from 'react';
import type { Role } from '../types';

interface ProjectRoleSelectorProps {
  roles: Role[];
  projectRoles: Role[];
  setProjectRoles: (roles: Role[]) => void;
  loading: boolean;
  error: string | null;
}

let uniqueIdCounter = 0;

const generateUniqueId = () => {
  return ++uniqueIdCounter;
};

const ProjectRoleSelector: React.FC<ProjectRoleSelectorProps> = ({
  roles,
  projectRoles,
  setProjectRoles,
  loading,
  error
}) => {
  const [selectedRoleId, setSelectedRoleId] = useState<number | string>('');
  const [percentage, setPercentage] = useState<number | string>('100');
  const [localError, setLocalError] = useState<string | null>(null);

  const handleAddRole = () => {
    setLocalError(null);
    
    if (!selectedRoleId) {
      setLocalError('Please select a role');
      return;
    }

    const roleId = Number(selectedRoleId);
    const percentageValue = Number(percentage);
    
    if (isNaN(percentageValue) || percentageValue <= 0 || percentageValue > 100) {
      setLocalError('Percentage must be between 1 and 100');
      return;
    }

    // Find the selected role to get its name
    const selectedRole = roles.find(r => r.id === roleId);
    if (!selectedRole) {
      setLocalError('Selected role not found');
      return;
    }

    // Add the new role with a unique identifier
    setProjectRoles([
      ...projectRoles,
      { 
        id: roleId, 
        name: selectedRole.name, 
        percentage: percentageValue,
        _uniqueId: generateUniqueId() // Use a robust unique identifier
      }
    ]);

    // Reset inputs
    setSelectedRoleId('');
    setPercentage('100');
  };

  // Use the _uniqueId to remove roles
  const handleRemoveRole = (uniqueId: number) => {
    setProjectRoles(projectRoles.filter(role => role._uniqueId !== uniqueId));
  };

  return (
    <div className="space-y-6">
      <h3 className="text-lg font-medium text-gray-900 dark:text-white">Project Roles</h3>
      
      {/* Role selection form */}
      <div className="flex flex-col md:flex-row gap-4 items-end">
        <div className="flex-grow">
          <label htmlFor="role-select" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Select Role
          </label>
          <select
            id="role-select"
            value={selectedRoleId}
            onChange={(e) => setSelectedRoleId(e.target.value)}
            disabled={loading}
            className="appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-700 focus:outline-none focus:ring-indigo-500 dark:focus:ring-indigo-600 focus:border-indigo-500 dark:focus:border-indigo-600 focus:z-10 text-sm"
          >
            <option value="">Select a role</option>
            {roles.map((role) => (
              <option 
                key={role.id} 
                value={role.id}
              >
                {role.name}
              </option>
            ))}
          </select>
        </div>
        
        <div className="w-full md:w-32">
          <label htmlFor="percentage-input" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Percentage (%)
          </label>
          <input
            id="percentage-input"
            type="number"
            min="1"
            max="100"
            value={percentage}
            onChange={(e) => setPercentage(e.target.value)}
            className="appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-700 focus:outline-none focus:ring-indigo-500 dark:focus:ring-indigo-600 focus:border-indigo-500 dark:focus:border-indigo-600 focus:z-10 text-sm"
          />
        </div>
        
        <div>
          <button
            type="button"
            onClick={handleAddRole}
            disabled={loading}
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
          >
            Add Role
          </button>
        </div>
      </div>
      
      {localError && (
        <p className="mt-2 text-sm text-red-600 dark:text-red-400">
          {localError}
        </p>
      )}

      {error && (
        <p className="mt-2 text-sm text-red-600 dark:text-red-400">
          {error}
        </p>
      )}
      
      {/* Roles table */}
      {projectRoles.length > 0 && (
        <div className="mt-6">
          <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">Assigned Roles</h4>
          <div className="bg-white dark:bg-gray-700 shadow overflow-hidden sm:rounded-md">
            <ul className="divide-y divide-gray-200 dark:divide-gray-600">
              {projectRoles.map((role) => (
                <li key={role._uniqueId || role.id} className="px-4 py-3 flex items-center justify-between text-sm">
                  <div className="flex items-center">
                    <span className="truncate font-medium text-gray-900 dark:text-white">
                      {role.name}
                    </span>
                    <span className="ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-800 dark:text-blue-100">
                      {role.percentage}%
                    </span>
                  </div>
                  <button
                    type="button"
                    onClick={() => handleRemoveRole(role._uniqueId || role.id)}
                    className="ml-4 flex-shrink-0 text-red-500 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
                  >
                    Remove
                  </button>
                </li>
              ))}
            </ul>
          </div>
        </div>
      )}
    </div>
  );
};

export default ProjectRoleSelector;