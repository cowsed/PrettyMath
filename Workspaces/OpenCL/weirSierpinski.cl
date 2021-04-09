long int fac(int n){
  long int prod = n;
  for (long int i=(long int)(n)-1; i>=1; i--){
    prod*=i;
  }
  
  return prod;
}

long int comb(int n, int r){
  return fac(n) / (fac(r)*fac(n-r));
}

__kernel void fractal(
  __write_only image2d_t image,
  const unsigned int cutoff,
  const float3 bgcol,
  const float3 fgcol,
  const float scale
) {
  int id = get_global_id(0);
  
  int2 size = get_image_dim(image);
  int Width = size.x;
  int Height = size.y;
  
  int idx = id % Width;
  int idy = (id / Width);
  float x = ((float)idx)*scale;
  float y = ((float)idy)*scale;

  int2 pos = (int2)((int)x,(int)y);  


  int c = (pos.x & pos.y)==0;
  
  float3 col=mix(bgcol,fgcol,(float)(c));
  
  write_imagef(image, (int2)(idx,idy), (float4)(col,1));
}

