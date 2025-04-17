import { useState } from "react";
import type { Route } from "./+types/login";
// Import useNavigate from the framework's router package
import { useNavigate } from "react-router";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Admin Login" },
    { name: "description", content: "Login to the admin section" },
  ];
}

export default function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<{ email?: string; password?: string; api?: string }>({});
  const [isLoading, setIsLoading] = useState(false); // Add loading state
  const navigate = useNavigate(); // Hook for navigation

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const newErrors: { email?: string; password?: string; api?: string } = {};

    // Basic validation (replace with more robust logic as needed)
    if (!email) {
      newErrors.email = "Email address is required.";
    } else if (!/\S+@\S+\.\S+/.test(email)) {
      // Basic email format check
      newErrors.email = "Please enter a valid email address.";
    }

    if (!password) {
      newErrors.password = "Password is required.";
    }

    setErrors(newErrors);

    if (Object.keys(newErrors).length === 0) {
      setIsLoading(true); // Set loading state
      setErrors({}); // Clear previous errors

      try {
        const response = await fetch("http://localhost:8080/api/v1/login", { // Use the backend endpoint
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ email, password }),
        });

        const data = await response.json();

        if (!response.ok) {
          // Handle backend errors (e.g., invalid credentials)
          setErrors({ api: data.error || `Login failed: ${response.statusText}` });
        } else {
          // Login successful
          console.log("Login successful, token:", data.token);
          // TODO: Store the token securely (e.g., localStorage, context, state management)
          localStorage.setItem('authToken', data.token); // Example: storing in localStorage
          // TODO: Redirect to a protected route or dashboard
          navigate("/dashboard"); // Example: redirect to dashboard
        }
      } catch (error) {
        // Handle network errors or unexpected issues
        console.error("Login request failed:", error);
        setErrors({ api: "Login request failed. Please try again later." });
      } finally {
        setIsLoading(false); // Reset loading state
      }
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-100 dark:bg-gray-900">
      <div className="w-full max-w-md rounded-lg bg-white p-8 shadow-md dark:bg-gray-800">
        <h2 className="mb-6 text-center text-2xl font-bold text-gray-900 dark:text-white">
          Admin Login
        </h2>
        <form onSubmit={handleSubmit} noValidate>
          <div className="mb-4">
            <label
              htmlFor="email"
              className="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
            >
              Email Address
            </label>
            <input
              type="email"
              id="email"
              name="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className={`block w-full rounded-md border-gray-300 px-3 py-2 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:focus:border-indigo-600 dark:focus:ring-indigo-600 sm:text-sm ${
                errors.email ? "border-red-500 dark:border-red-500" : ""
              }`}
              placeholder="you@example.com"
              aria-invalid={errors.email ? "true" : "false"}
              aria-describedby={errors.email ? "email-error" : undefined}
            />
            {errors.email && (
              <p id="email-error" className="mt-1 text-sm text-red-600 dark:text-red-400">
                {errors.email}
              </p>
            )}
          </div>
          <div className="mb-6">
            <label
              htmlFor="password"
              className="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
            >
              Password
            </label>
            <input
              type="password"
              id="password"
              name="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className={`block w-full rounded-md border-gray-300 px-3 py-2 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:focus:border-indigo-600 dark:focus:ring-indigo-600 sm:text-sm ${
                errors.password ? "border-red-500 dark:border-red-500" : ""
              }`}
              placeholder="••••••••"
              aria-invalid={errors.password ? "true" : "false"}
              aria-describedby={errors.password ? "password-error" : undefined}
            />
            {errors.password && (
              <p id="password-error" className="mt-1 text-sm text-red-600 dark:text-red-400">
                {errors.password}
              </p>
            )}
          </div>
          {errors.api && (
            <p id="api-error" className="mb-4 text-sm text-red-600 dark:text-red-400 text-center">
              {errors.api}
            </p>
          )}
          <div>
            <button
              type="submit"
              disabled={isLoading} // Disable button when loading
              className="flex w-full justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:opacity-50 dark:focus:ring-offset-gray-800"
            >
              {isLoading ? "Signing in..." : "Sign in"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
