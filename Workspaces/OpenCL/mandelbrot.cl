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
  return newCmplx(pow((pow(a,2.0f)+pow(b,2.0f)),(n/2) )*(cos(n *atan(b/a))), 
  pow((pow(a,2.0f)+pow(b,2)),(n/2) )*(sin(n *atan(b/a))));
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

bool equals(cmplx a, cmplx b){
  return a.real==b.real&&a.imag==b.imag;
}
bool equalsApprox(cmplx a, cmplx b, float epsilon){
  return fabs(a.real-b.real)<=epsilon&&fabs(a.imag-b.imag)<epsilon;
}

int checkIfInLog(cmplx orbitLog[40], int logLen, cmplx a, float epsilon){
  for (int i=0; i<logLen; i++){
    if (equalsApprox(a,orbitLog[i],epsilon)){
      return logLen-i;
    }
  }
  //Not found
  return -1;
}
float iterate(int max_iteration, float radius, float x, float y, float n, float epsilon){

  int id=get_global_id(0);
  cmplx c = newCmplx(x,y);
  cmplx z = newCmplx(x,y);
  
  
  cmplx orbitLog[40];
  int orbitLogLength = 0;
  
  int iteration = 0;

  float amt=0;

  while (clen(z) < radius&& iteration < max_iteration){
    cmplx zp = cpow(z,n);
    z=cadd(zp,c);


    iteration++;
    
    amt=(((float)iteration)/(float)max_iteration);
    if (iteration == max_iteration){ // Belongs to the set
      break;
    } 
    int period=checkIfInLog(orbitLog, orbitLogLength, z, epsilon);
    if (period!=-1){
      iteration=(int)((float)period/20.0);
      break;
    } else if(orbitLogLength<40){
      orbitLog[orbitLogLength]=z;
      orbitLogLength++;
    }
  }
  return amt;
}
__kernel void fractal(
  __write_only image2d_t image,
  const float n,
  const float2 pos,
  const float scaleInv,
  const unsigned int iterations,
  const float radius,
  const float3 bgcol,
  const float3 fgcol,
  const unsigned int SSAmtI,
  const float epsilon
) {
  int id = get_global_id(0);
  
  float posx=pos.x;
  float posy=pos.y;
  
  float scale = 16.0/scaleInv;
  float SSAmt = (float)(SSAmtI);

  int2 size = get_image_dim(image);
  int Width = size.x;
  int Height = size.y;
  

  int idx = id % Width;
  int idy = (id / Width);


  float x0 = ((float)idx)/(float)Width;
  float y0 = ((float)idy)/(float)Height;
  float3 sumCol = (float3)(0,0,0);

  for (float offY = -SSAmt/2; offY<SSAmt/2; offY++){
    for (float offX = -SSAmt/2; offX<SSAmt/2; offX++){
      float x = x0+(offX/SSAmt)/(float)Width;
      float y = y0+(offY/SSAmt)/(float)Height;
      float2 uv = (float2)(x,y)-(float2)(0.5);
  
      float2 pos2 = uv*scale + pos;

      float amt =iterate(iterations,radius, pos2.x,pos2.y,n,epsilon);
      float3 col = mix(bgcol, fgcol, amt);
      sumCol+=col;
    }
  }

  sumCol/=(float)(SSAmt*SSAmt);

  write_imagef(image, (int2)(idx,idy), (float4)(sumCol,1));
}

