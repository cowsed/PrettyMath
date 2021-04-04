typedef struct cmplx {
  float real;
  float imag;
} cmplx;

cmplx newCmplx(float real, float imag){
  cmplx c;
  c.real = real;
  c.imag = imag;
  return c;
}

float real(cmplx a){
  return a.real;
}
float imag(cmplx a){
  return a.imag;
}

cmplx cpow(cmplx c, float n){
  float a = real(c);
  float b = imag(c);
  return newCmplx(pow((pow(a,2.0f)+pow(b,2.0f)),(n/2) )*(cos(n *atan(b/a))), pow((pow(a,2.0f)+pow(b,2)),(n/2) )*(sin(n *atan(b/a))));
}


cmplx cadd(cmplx a, cmplx b){
  float aa,ab,ba,bb;
  aa=real(a);
  ab=imag(a);
  ba=real(b);
  bb=imag(b);
  cmplx c = newCmplx(aa+ba, ab+bb);
  return c;
}

float clen(cmplx a){
  float x = real(a);
  float y = imag(a);
  return x*x+y*y;
}
cmplx csqr(cmplx a){
  float x = real(a)*real(a) - imag(a)*imag(a);
  float y = 2*a.real*a.imag;
  return newCmplx(x,y);
}
float iterate(float x, float y, float n){
  float zx = x; // zx represents the real part of z
  float zy = y; // zy represents the imaginary part of z 
  
  int id=get_global_id(0);
  cmplx c = newCmplx(x,y);
  cmplx z = newCmplx(x,y);
  
  if (id ==0 ){
    printf("c: real: %.2f%+.2fi  |  %.2f%+.2fi\n",x,y,c.real, c.imag);
    printf("z0: real: %.2f%+.2fi  |  %.2f%+.2fi\n",zx,zy,z.real, z.imag);
  }
  
  int iteration = 0;
  int max_iteration = 64;
  float radius = 60;
  float amt;

  while (clen(z) < radius&& iteration < max_iteration){
    //float xtemp = pow((float)zx,n) - pow((float)zy,n) + x;

    //a^3-ab^2+a^2bi-2ab^2
    //2a^2bi-b^3i
    //float xtemp = pow(zx,3)-zx*pow(zy,2)-2*zx*pow(zy,2)+x;
    
    float xtemp = zx*zx - zy*zy + x;
    zy=(2*zx*zy)+y;
    zx=xtemp;
    
    cmplx zp = cpow(z,2);
    z=cadd(zp,c);

     if (id ==0 ){
      printf("after iter: %d, correct: %.2f%+.2fi  |  other: %.2f%+.2fi\n", iteration, zx,zy, z.real, z.imag);
    }
    iteration++;
    
    amt=(((float)iteration)/(float)max_iteration);
    if (iteration == max_iteration){ // Belongs to the set
      break;
    } 
  }
  return amt;
}
__kernel void blur(
  __write_only image2d_t image,
  const float n,
  const float posx,
  const float posy,
  const float scale,
  const float SSAmt,
  const float bgr,
  const float bgg,
  const float bgb
) {
  int id = get_global_id(0);
  
  if(id==0){
    printf("==am 0==\n");
    cmplx c = newCmplx(1,2);
    printf("c=%.1f+%.1fi \n",c.real, c.imag);
    printf("c=%.1f+%.1fi \n",real(c), imag(c));
    cmplx c2 = cadd(c, c);
    printf("c+c=%.1f+%.1fi\n", c2.real, c2.imag);
    cmplx cs1 = csqr(c);
    printf("A. c^2=%.1f+%.1fi\n", cs1.real, cs1.imag);
    cmplx cs2 = cpow(c,2);
    printf("B. c^2=%.1f+%.1fi\n", cs2.real, cs2.imag);

    }

  float3 bgcol = (float3)(bgr,bgg,bgb);
  float3 fgcol = (float3)(1,1,1);


  int2 size = get_image_dim(image);
  int Width = size.x;
  int Height = size.y;
  

  int idx = id % Width;
  int idy = (id / Width);


  float x0 = ((float)idx)/(float)Width;
  float y0 = ((float)idy)/(float)Height;
  //UV is now [-.5,.5] and [-.5,.5]
  float3 sumCol = (float3)(0,0,0);//(float4)(uv.x,uv.y,0.0,1);

  for (float offY = -SSAmt/2; offY<SSAmt/2; offY++){
    for (float offX = -SSAmt/2; offX<SSAmt/2; offX++){
      float x = x0+(offX/SSAmt)/(float)Width;
      float y = y0+(offY/SSAmt)/(float)Height;
      float2 uv = (float2)(x,y)-(float2)(0.5);
  

      x = uv.x*scale+posx; //scaled x coordinate of pixel (scaled to lie in the Mandelbrot X scale (-2.5, 1))
      y = uv.y*scale+posy; //scaled y coordinate of pixel (scaled to lie in the Mandelbrot Y scale (-1, 1))
      float amt =iterate(x,y,n);
      float3 col = mix(bgcol, fgcol, amt);
      sumCol+=col;
    }
  }

  sumCol/=(float)(SSAmt*SSAmt);

  write_imagef(image, (int2)(idx,idy), (float4)(sumCol,1));
}

