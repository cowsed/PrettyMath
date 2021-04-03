__kernel void blur(
  __write_only image2d_t image,
  const float posx,
  const float posy,
  const float scale
) {


  int2 size = get_image_dim(image);
  int Width = size.x;
  int Height = size.y;
  
  int id = get_global_id(0);
  int idx = id % Width;
  int idy = (id / Width);

  float SSAmt = 16.0;

  float x0 = ((float)idx)/(float)Width;
  float y0 = ((float)idy)/(float)Height;
  //UV is now [-.5,.5] and [-.5,.5]
  float3 sumCol = (float3)(0,0,0);//(float4)(uv.x,uv.y,0.0,1);

  for (float offY = -SSAmt/2; offY<SSAmt/2; offY++){
    for (float offX = -SSAmt/2; offX<SSAmt/2; offX++){
      float x = x0+(offX/SSAmt)/(float)Width;
      float y = y0+(offY/SSAmt)/(float)Height;
      if (idx==1000&&idy==600){printf("(%f,%f)",x,y);}
      float2 uv = (float2)(x,y)-(float2)(0.5);
  

      float3 col = (float3)(0,0,.4);//(float4)(uv.x,uv.y,0.0,1);
      x = uv.x*scale+posx; //scaled x coordinate of pixel (scaled to lie in the Mandelbrot X scale (-2.5, 1))
      y = uv.y*scale+posy; //scaled y coordinate of pixel (scaled to lie in the Mandelbrot Y scale (-1, 1))

      float zx = x; // zx represents the real part of z
      float zy = y; // zy represents the imaginary part of z 

      int iteration = 0;
      int max_iteration = 64;

      while (zx*zx + zy*zy < 4 && iteration < max_iteration){
        float xtemp = zx*zx - zy*zy + x;
        zy = fabs(2*zx*zy) + y; // abs returns the absolute value
        zx = xtemp;
        iteration++;
        if (iteration == max_iteration){ // Belongs to the set
          float amt=(((float)iteration)/(float)max_iteration);

          col=(float3)(amt,amt,amt);
          break;
        }
      }
      sumCol+=col;
    }
  }

  sumCol/=(float)(SSAmt*SSAmt);

  write_imagef(image, (int2)(idx,idy), (float4)(sumCol,1));
}
