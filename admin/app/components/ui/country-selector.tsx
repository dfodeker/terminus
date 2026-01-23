'use client';
import { countries } from "@/constants";
import { useState } from "react";


interface CountrySelectProps {
  defaultCountry?: string;
  value?: string;
  onChange?: (code: string) => void;
}

export function CountrySelect({ defaultCountry = 'CA', value, onChange }: CountrySelectProps) {
  const [selected, setSelected] = useState(value ?? defaultCountry);

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const code = e.target.value;
    setSelected(code);
    onChange?.(code);
  };

  return (
    <select
      value={selected}
      onChange={handleChange}
      className="text-sm text-gray-700 bg-transparent border border-gray-200 rounded-md px-2 py-1 cursor-pointer hover:border-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-200"
    >
      {countries.map((country) => (
        <option key={country.code} value={country.code} className="flex space-x-2 items-center gap-2">
            {country.emoji} {country.name}
            
        </option>
      ))}
    </select>
  );
}