interface FormMessageProps {
  message: string;
  success: boolean;
}

export function FormMessage({ message, success }: FormMessageProps) {
  if (!message) return null;

  const styles = success
    ? 'bg-green-50 border-green-200 text-green-600'
    : 'bg-red-50 border-red-200 text-red-600';

  return (
    <div className={`p-3 rounded-md border ${styles}`}>
      <p className="text-sm">{message}</p>
    </div>
  );
}
