                                                                                                                                                            
  import { getAuthenticatedUser } from '@/lib/auth';
                                                                                                                   
                                                                                                                                                                             
  export async function UserGreetingServer() {                                                                                                                                      
    const user = await getAuthenticatedUser();                                                                                                                               
    return <p>Welcome, {user.email}</p>;                                                                                                                                     
  }    