import { Suspense } from 'react';
import { UserGreetingServer } from './hi';

export default function Page(){
    return (
    <div>
      <h1 className="text-2xl font-bold">Admin Dashboard</h1>
      <p>If you see this at admin.localhost:3000, subdomain routing works!</p>
      <Suspense fallback={<p>Loading...</p>}>                                                                                                                              
          <UserGreetingServer />                                                                                                                                             
        </Suspense> 
    </div>
  );
}